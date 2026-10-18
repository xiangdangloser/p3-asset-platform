package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiangdangloser/p3-asset-platform/database"
	"github.com/xiangdangloser/p3-asset-platform/models"
)

// GetAssets 获取所有资产列表
func GetAssets(c *gin.Context) {
	// ⭐ 获取当前用户 ID
	userID := c.GetUint("user_id")

	var assets []models.Asset
	// ⭐ 过滤查询
	database.DB.Where("user_id = ?", userID).Find(&assets)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": assets,
	})
}

// CreateAsset 新增一个资产
func CreateAsset(c *gin.Context) {
	var newAsset models.Asset
	if err := c.ShouldBindJSON(&newAsset); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 写入数据库
	// 自动设置 createdAt 字段
	newAsset.CreatedAt = time.Now()
	if result := database.DB.Create(&newAsset); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "写入数据库失败：" + result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "资产新增成功",
		"data":    newAsset,
	})
}
