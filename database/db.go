package database

import (
	"log"
	"os"
	"time"

	"github.com/xiangdangloser/p3-asset-platform/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// 从环境变量读取 DSN，如果读不到就用默认的 SQLite（方便本地开发）
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// 如果没配环境变量，默认走 MySQL 的本地测试端口（也可以切回 sqlite）
		dsn = "user:userpassword@tcp(127.0.0.1:3306)/asset_db?charset=utf8mb4&parseTime=True&loc=Local"
	}

	var err error
	// 重试连接 MySQL（应对 Docker 启动时序问题）
	for i := 0; i < 10; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("等待 MySQL 启动中 (第 %d 次重试)... 错误: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("数据库连接失败，请检查 MySQL 状态: %v", err)
	}

	// 自动迁移表结构
	err = DB.AutoMigrate(&models.Asset{}, &models.Task{}, &models.User{})
	if err != nil {
		log.Fatalf("数据表迁移失败: %v", err)
	}

	log.Println("数据库初始化成功，已连接 MySQL")
}
