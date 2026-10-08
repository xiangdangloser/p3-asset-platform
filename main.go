package main

import (
	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/routes"
)

func main() {
	// 1. 初始化数据库
	database.InitDB()

	// 获取配置好的路由
	r := routes.SetupRouter()

	// 3. 启动服务，监听本地 8080 端口
	r.Run(":8080")
}
