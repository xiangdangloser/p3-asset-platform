package models

import "time"

type Asset struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"index" json:"user_id"` // ⭐ 新增
	// ⭐ 改为复合唯一索引
	Target      string    `gorm:"type:varchar(255);uniqueIndex:idx_user_target" json:"target"`
	StatusCode  int       `json:"status_code"`
	Server      string    `json:"server"`
	Title       string    `json:"title"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}
