package models

import "time"

// Task 代表一个后台异步扫描任务
type Task struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"` // ⭐ 新增
	Status    string    `json:"status"`               // 状态: running, completed, failed
	Total     int       `json:"total"`                // 目标总数
	CreatedAt time.Time `json:"created_at"`
}
