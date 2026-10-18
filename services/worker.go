package services

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/models"
	"gorm.io/gorm/clause"
)

type ScanTaskPayload struct {
	TaskID  uint     `json:"task_id"`
	UserID  uint     `json:"user_id"` // ⭐ 新增
	Targets []string `json:"targets"`
}

const (
	QueuePending    = "scan_queue"
	QueueProcessing = "scan_processing" // ⭐ 新增：处理中队列
)

func StartWorker(ctx context.Context) {
	log.Println("Worker 启动，开始监听 Redis 队列...")

	// 1. ⭐ 启动时，先恢复上次中断的任务（断点恢复）
	recoverProcessingTasks(ctx)

	// 2. 无限循环消费队列
	for {
		// 检查上下文是否被取消（优雅停机）
		select {
		case <-ctx.Done():
			log.Println("Worker 收到停止信号，正在退出...")
			return
		default:
		}

		// ⭐ 使用 BRPopLPush：从 Pending 取出，同时原子地放入 Processing
		// 如果 Worker 崩溃，消息依然在 Processing 队列里
		//
		// ⚠️ 超时必须是有限值（1s），不能是 0（永久阻塞）：
		// go-redis 的 ctx 取消只作用于「带 Deadline 的 ctx」，而 signal.NotifyContext
		// 的 ctx 只有取消没有 Deadline，socket 读会永久阻塞，停机信号永远打不进来了
		// （实测 docker stop 因此等了 15s 被强杀）。1s 超时只是空转一次，几乎无开销。
		result, err := database.RDB.BRPopLPush(ctx, QueuePending, QueueProcessing, time.Second).Result()
		if err != nil {
			// ⭐ 收到停机信号时 BRPopLPush 会立即返回 context.Canceled，安静退出即可，
			// 不需要打错误日志（之前的版本会误报"拉取失败"还白等 2 秒）
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("Worker 收到停止信号，正在退出...")
				return
			}
			// 队列为空：1s 超时正常空转，回到循环顶部检查停机信号
			if err == redis.Nil {
				continue
			}
			log.Printf("拉取 Redis 任务失败: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var payload ScanTaskPayload
		if err := json.Unmarshal([]byte(result), &payload); err != nil {
			log.Printf("解析任务数据失败: %v", err)
			// 脏数据直接丢弃，防止死循环
			database.RDB.LRem(ctx, QueueProcessing, 1, result)
			continue
		}

		log.Printf("开始处理任务 ID: %d, 目标数: %d", payload.TaskID, len(payload.Targets))
		// ⭐ 传入 payload.UserID
		processScan(ctx, payload.TaskID, payload.UserID, payload.Targets)
		// ⭐ ACK：只有任务正常跑完才从 Processing 队列确认删除。
		// 若被停机信号中断（ctx 已取消），故意不 ACK，消息留在 Processing 队列，
		// 下次启动时 recoverProcessingTasks 会把它重新入队重扫，保证不丢任务。
		if ctx.Err() == nil {
			database.RDB.LRem(ctx, QueueProcessing, 1, result)
		}
	}
}

// 启动时，把上次没跑完的任务从 processing_queue 挪回 pending 队列
func recoverProcessingTasks(ctx context.Context) {
	for {
		// RPopLPush：从 Processing 右侧弹出，放回 Pending 左侧，保证不丢
		val, err := database.RDB.RPopLPush(ctx, QueueProcessing, QueuePending).Result()
		if err == redis.Nil {
			break // 没有孤儿任务了，退出循环
		}
		if err != nil {
			log.Printf("恢复孤儿任务失败: %v", err)
			break
		}
		log.Printf("发现未完成的任务，重新入队: %s", val)
	}
}

// processScan 执行实际扫描逻辑，并增加 Panic 兜底
func processScan(ctx context.Context, taskID uint, userID uint, targets []string) {
	// ⭐ 新增：Panic 兜底，防止一个坏任务导致整个 Worker 崩溃
	defer func() {
		if r := recover(); r != nil {
			log.Printf("任务 ID: %d 执行时发生崩溃: %v", taskID, r)
			// 标记任务为 failed
			database.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("status", "failed")
		}
	}()

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
				// ⭐ 优雅停机：收到取消信号后不再发起新请求，剩余目标留给下次启动恢复重扫
				if ctx.Err() != nil {
					continue
				}
				if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
					target = "http://" + target
				}
				status, server, title, fp, err := Probe(client, target, 2)
				asset := models.Asset{Target: target, CreatedAt: time.Now(), UserID: userID}
				if err != nil {
					asset.StatusCode = 0
					asset.Server, asset.Title, asset.Fingerprint = "-", "-", "-"
				} else {
					asset.StatusCode, asset.Server, asset.Title, asset.Fingerprint = status, server, title, fp
				}
				mu.Lock()
				database.DB.Clauses(clause.OnConflict{
					// ⭐ 冲突列变为 UserID + Target
					Columns:   []clause.Column{{Name: "user_id"}, {Name: "target"}},
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

	// ⭐ 若任务被停机信号中断，改回 pending（消息未 ACK，下次启动会重新入队重扫），
	// 绝不能标记 completed，否则前端会误以为全部扫完了
	if ctx.Err() != nil {
		database.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("status", "pending")
		log.Printf("任务 ID: %d 被停机信号中断，已标记为 pending，等待下次启动恢复", taskID)
		return
	}

	database.DB.Model(&models.Task{}).Where("id = ?", taskID).Update("status", "completed")
	log.Printf("任务 ID: %d 执行完毕", taskID)
}
