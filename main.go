package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/routes"
	"github.com/xiangdangloser/p3-asset-platform/services"
)

func main() {
	// 1. 初始化MySql
	database.InitDB()

	// 2. 初始化 Redis
	database.InitRedis()

	// 2. ⭐ 创建一个监听中断信号的 Context（优雅停机核心）
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. 启动后台扫描 Worker (异步)
	go services.StartWorker(ctx)

	// 获取配置好的路由
	r := routes.SetupRouter()
	srv := &http.Server{Addr: ":8080", Handler: r}

	go func() {
		log.Println("API 服务启动，监听 :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API 服务启动失败: %v", err)
		}
	}()

	// 5. ⭐ 主协程阻塞，等待中断信号
	<-ctx.Done()
	log.Println("收到停机信号，正在优雅关闭服务（等待正在进行的扫描收尾）...")

	// 6. 给正在跑的任务 30 秒的收尾时间
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("服务强制关闭: %v", err)
	}
	log.Println("服务已安全退出")

}
