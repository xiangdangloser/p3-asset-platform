package database

import (
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"github.com/xiangdangloser/p3-asset-platform/models"
	"gorm.io/gorm"
)

// DB 是一个全局数据库连接对象
var DB *gorm.DB

// InitDB 初始化数据库连接并自动建表
func InitDB() {
	// 1. 动态获取数据库路径
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "assets.db" // 默认本地开发路径
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动迁移：根据 models.Asset 结构体自动创建数据表
	err = DB.AutoMigrate(&models.Asset{}, &models.Task{}, &models.User{})
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}
	log.Println("数据库初始化成功，已连接 assets.db (纯 Go 驱动)")
}
