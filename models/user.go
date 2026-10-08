package models

import "time"

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50" json:"username"`
	Password  string    `json:"-"` // json:"-" 表示返回JSON数据时不输出密码字段
	CreatedAt time.Time `json:"created_at"`
}
