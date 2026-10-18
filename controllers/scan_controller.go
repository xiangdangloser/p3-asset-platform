package controllers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/models"
	"github.com/xiangdangloser/p3-asset-platform/services"
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

	// ⭐ 获取当前用户 ID
	userID := c.GetUint("user_id")

	log.Printf("调试: 当前用户ID为 %d", userID) // ⭐ 加一行日志，立刻看出是不是 0

	// 1. 创建数据库任务记录（状态为 pending）
	task := models.Task{Status: "pending", Total: len(req.Targets), UserID: userID}
	database.DB.Create(&task)

	// 2. 封装任务数据
	payload := services.ScanTaskPayload{
		TaskID:  task.ID,
		UserID:  userID,
		Targets: req.Targets,
	}
	payloadBytes, _ := json.Marshal(payload)

	// 3. 推入 Redis 队列 (LPUSH)
	if err := database.RDB.LPush(context.Background(), "scan_queue", payloadBytes).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "任务入队失败: " + err.Error()})
		return
	}

	// 4. 立刻返回 task_id
	c.JSON(http.StatusOK, gin.H{
		"message": "任务已入队，等待后台执行",
		"task_id": task.ID,
		"user_id": userID,
	})
}
