package models

import "time"

type Asset struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Target      string    `gorm:"type:varchar(255);uniqueIndex" json:"target"` // 建立唯一索引，防止重复扫同一个目标
	StatusCode  int       `json:"status_code"`
	Server      string    `json:"server"`
	Title       string    `json:"title"`
	Fingerprint string    `json:"fingerprint"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
}
