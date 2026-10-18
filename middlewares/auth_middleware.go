package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("xiangdangloser_secret_key") // 必须和控制器里的密钥一致

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// ⚠️ 修复：这里是 || 不是 &&。
		// 原代码用 && 时，"Basic xxx" 这类非 Bearer 开头的 Header 会绕过本判断直接进入解析流程。
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供 Token 或格式错误"})
			c.Abort()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		// 限定签名方法为 HS256，防止算法混淆（alg confusion）攻击
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token 无效或已过期"})
			c.Abort()
			return
		}
		// ⭐ 新增：提取 user_id 并存入上下文
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// JWT 里的数字默认是 float64，需要转成 uint
			if userIDFloat, ok := claims["user_id"].(float64); ok {
				c.Set("user_id", uint(userIDFloat))
			}
		}

		c.Next() // 放行

	}
}
