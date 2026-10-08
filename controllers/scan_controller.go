package controllers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/models"
	"github.com/xiangdangloser/p3-asset-platform/services"
	"gorm.io/gorm/clause"
)

type ScanRequest struct {
	Targets []string `json:"targets"`
}

func ScanTargets(c *gin.Context) {
	var req ScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误: " + err.Error()})
		return
	}

	// 1. 在数据库里建一个任务记录（Task 模型）
	task := models.Task{
		Status: "running",
		Total:  len(req.Targets),
	}
	database.DB.Create(&task)

	// ========================================================
	// 2. ⚠️ 核心：立刻返回 HTTP 响应给用户！绝不等待！
	// ========================================================
	c.JSON(http.StatusOK, gin.H{
		"message": "任务已下发，正在后台扫描",
		"task_id": task.ID,
	})

	// ========================================================
	// 3. 所有扫描逻辑，全部包裹在这个后台 goroutine 里
	// ========================================================
	go func(taskID uint, targets []string) {
		client := &http.Client{Timeout: 5 * time.Second}
		var mu sync.Mutex

		jobs := make(chan string, len(targets))
		var wg sync.WaitGroup

		// 启动 10 个并发 Worker
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for target := range jobs {
					if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
						target = "http://" + target
					}

					status, server, title, fp, err := services.Probe(client, target, 2)

					asset := models.Asset{
						Target:    target,
						CreatedAt: time.Now(),
					}

					if err != nil {
						asset.StatusCode = 0
						asset.Server = "-"
						asset.Title = "-"
						asset.Fingerprint = "-"
					} else {
						asset.StatusCode = status
						asset.Server = server
						asset.Title = title
						asset.Fingerprint = fp
					}

					// 加锁写入数据库，并处理重复目标的覆盖更新
					mu.Lock()
					database.DB.Clauses(clause.OnConflict{
						Columns:   []clause.Column{{Name: "target"}},
						DoUpdates: clause.AssignmentColumns([]string{"status_code", "server", "title", "fingerprint", "created_at"}),
					}).Create(&asset)
					mu.Unlock()
				}
			}()
		}

		// 生产者派发任务
		for _, t := range targets {
			jobs <- t
		}
		close(jobs)

		// 等待所有任务跑完
		wg.Wait()

		// 4. 任务跑完，把数据库里 Task 状态改为 completed
		database.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("status", "completed")
	}(task.ID, req.Targets)
}
