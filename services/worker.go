package services

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/models"
	"gorm.io/gorm/clause"
)

type ScanTaskPayload struct {
	TaskID  uint     `json:"task_id"`
	Targets []string `json:"targets"`
}

func StartWorker() {
	log.Println("Worker 启动，开始监听 Redis 队列...")
	ctx := context.Background()

	for {
		// 使用最简单的 BRPOP，稳定可靠，不会因为阻塞机制导致卡死
		result, err := database.RDB.BRPop(ctx, 0, "scan_queue").Result()
		if err != nil {
			log.Printf("拉取 Redis 任务失败: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var payload ScanTaskPayload
		if err := json.Unmarshal([]byte(result[1]), &payload); err != nil {
			continue
		}

		log.Printf("开始处理任务 ID: %d, 目标数: %d", payload.TaskID, len(payload.Targets))
		processScan(payload.TaskID, payload.Targets)
	}
}

func processScan(taskID uint, targets []string) {
	database.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("status", "running")

	client := &http.Client{Timeout: 5 * time.Second}
	var mu sync.Mutex
	jobs := make(chan string, len(targets))
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for target := range jobs {
				if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
					target = "http://" + target
				}
				status, server, title, fp, err := Probe(client, target, 2)
				asset := models.Asset{Target: target, CreatedAt: time.Now()}
				if err != nil {
					asset.StatusCode = 0
					asset.Server, asset.Title, asset.Fingerprint = "-", "-", "-"
				} else {
					asset.StatusCode, asset.Server, asset.Title, asset.Fingerprint = status, server, title, fp
				}
				mu.Lock()
				database.DB.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "target"}},
					DoUpdates: clause.AssignmentColumns([]string{"status_code", "server", "title", "fingerprint", "created_at"}),
				}).Create(&asset)
				mu.Unlock()
			}
		}()
	}

	for _, t := range targets {
		jobs <- t
	}
	close(jobs)
	wg.Wait()

	database.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("status", "completed")
	log.Printf("任务 ID: %d 执行完毕", taskID)
}
