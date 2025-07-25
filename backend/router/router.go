package router

import (
	"github.com/gin-gonic/gin"

	"movingManager/controller"
	"movingManager/middleware"
)

// RegisterRoutes 注册所有路由
func RegisterRoutes(r *gin.Engine) {
	// 公开路由组(无需认证)
	public := r.Group("/api/v1")
	{
		// 用户注册/登录
		public.POST("/user/auth", controller.UserAuth)
	}

	// 需要认证的路由组
	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware()) // 应用认证中间件
	{
		// 搬运模块
		move := api.Group("/move")
		{
			move.POST("/create", controller.CreateMove)       // 创建搬运
			move.POST("/detail", controller.GetMoveDetail)    // 搬运详情
			move.POST("/update", controller.UpdateMove)       // 编辑搬运
			move.POST("/delete", controller.DeleteMove)       // 删除搬运
			move.POST("/list", controller.GetMoveList)        // 搬运列表
		}

		// 标签模块
		 tag := api.Group("/tag")
		{
			tag.POST("/create", controller.CreateTag)         // 创建标签
			tag.POST("/update", controller.UpdateTag)         // 编辑标签
			tag.POST("/delete", controller.DeleteTag)         // 删除标签
			tag.POST("/verify", controller.VerifyTag)         // 核销标签
			tag.POST("/detail", controller.GetTagDetail)      // 标签详情
			tag.POST("/list", controller.GetTagList)          // 标签列表
			tag.POST("/generate-pdf", controller.GeneratePDF) // 生成PDF
		}
	}
}