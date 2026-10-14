package main

import (
	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/routes"
	"github.com/xiangdangloser/p3-asset-platform/services"
)

func main() {
	// 1. 初始化MySql
	database.InitDB()

	// 2. 初始化 Redis
	database.InitRedis()

	// 3. 启动后台扫描 Worker (异步)
	go services.StartWorker()

	// 获取配置好的路由
	r := routes.SetupRouter()

	// 3. 启动服务，监听本地 8080 端口
	r.Run(":8080")
}
