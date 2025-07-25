package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"movingManager/common"
	"movingManager/database"
	"movingManager/model"
)

// AuthMiddleware 认证中间件，验证用户登录状态
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization请求头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeUserNotLogin,
				"message": common.CodeMessage[common.CodeUserNotLogin],
			})
			c.Abort()
			return
		}

		// 验证token格式 (Bearer token)
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if !(len(tokenParts) == 2 && tokenParts[0] == "Bearer") {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeUserNotLogin,
				"message": common.CodeMessage[common.CodeUserNotLogin],
			})
			c.Abort()
			return
		}

		// 查询用户是否存在
		var user model.UserModel
		db := database.DB
		if err := db.Where("authorization_code = ? AND is_deleted = 0", tokenParts[1]).First(&user).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeUserNotRegistered,
				"message": common.CodeMessage[common.CodeUserNotRegistered],
			})
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("userUid", user.UserUid)
		c.Set("userName", user.UserName)

		c.Next()
	}
}
