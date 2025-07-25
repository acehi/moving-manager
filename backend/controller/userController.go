package controller

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"movingManager/common"
	"movingManager/service"
)

// UserAuthRequest 用户登录/注册请求参数
type UserAuthRequest struct {
	Mobile string `json:"mobile" binding:"required"`
}

// UserAuth 用户注册/登录接口
// 处理用户手机号登录，已存在则返回token，不存在则创建新用户
func UserAuth(c *gin.Context) {
	var req UserAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证手机号格式
	if !isValidMobile(req.Mobile) {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "手机号格式不正确",
		})
		return
	}

	// 调用服务层处理业务逻辑
	user, token, err := service.UserAuth(req.Mobile)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "操作失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": common.CodeMessage[common.CodeSuccess],
		"user": gin.H{
			"user_name": user.UserName,
		},
		"token": token,
	})
}

// isValidMobile 验证手机号格式
func isValidMobile(mobile string) bool {
	// 简单手机号正则验证(11位数字)
	pattern := `^1[3-9]\d{9}$`
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(mobile)
}
