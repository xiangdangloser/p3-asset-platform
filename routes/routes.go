package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiangdangloser/p3-asset-platform/controllers"
	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/middlewares"
	"github.com/xiangdangloser/p3-asset-platform/models"
)

// SetupRouter 初始化并配置路由
func SetupRouter() *gin.Engine {
	//  创建一个默认的 Gin 引擎（包含日志和恢复中间件）
	r := gin.Default()

	// 1. 加载 HTML 模板
	r.LoadHTMLGlob("templates/*")

	// 2. 配置静态文件服务
	r.Static("/static", "./static")

	// 3. 新增：渲染登录页
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	// 新增：Dashboard 页面（前端可以自己检查 Token，后端只负责返回 HTML）
	r.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", nil)
	})

	// 创建一个 API 分组，规范版本管理
	apiV1 := r.Group("/api/v1")
	{
		// 1. 公开接口（无需鉴权）
		apiV1.POST("/register", controllers.Register)
		apiV1.POST("/login", controllers.Login)
		// 资产相关的路由
		// 2. 受保护接口（需要 Token）
		protected := apiV1.Group("")
		protected.Use(middlewares.JWTAuth()) // 应用 JWT 中间件
		{
			assetGroup := protected.Group("/assets")
			{
				assetGroup.GET("", controllers.GetAssets)    // GET /api/v1/assets
				assetGroup.POST("", controllers.CreateAsset) // POST /api/v1/assets
			}
			// 新增扫描路由
			protected.POST("/scan", controllers.ScanTargets)

			// ==========================================
			// ⚠️ 新增：查询任务状态的路由
			// ==========================================
			protected.GET("/tasks/:id", func(c *gin.Context) {
				var task models.Task
				// 从 URL 中获取 id 参数并查询数据库
				if err := database.DB.First(&task, c.Param("id")).Error; err != nil {
					c.JSON(http.StatusNotFound, gin.H{"error": "任务不存在"})
					return
				}
				c.JSON(http.StatusOK, gin.H{"data": task})
			})
		}

	}
	return r
}
