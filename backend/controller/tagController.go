package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"movingManager/common"
	"movingManager/service"
)

// CreateTagRequest 创建标签请求参数
type CreateTagRequest struct {
	MoveUid string `json:"move_uid" binding:"required,uuid"`    // 搬运UID
	TagName string `json:"tag_name" binding:"required,max=100"` // 标签名称
	Remark  string `json:"remark" binding:"max=500"`            // 标签备注
	Status  int    `json:"status" binding:"omitempty,oneof=0 1 2"` // 标签状态(0-正常,1-锁定,2-已完成)
}

// CreateTag 创建标签接口
func CreateTag(c *gin.Context) {
	var req CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户UID
	userUid, exists := c.Get("userUid")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    common.CodeUserNotLogin,
			"message": common.CodeMessage[common.CodeUserNotLogin],
		})
		return
	}

	// 转换为服务层请求结构体
	serviceReq := service.CreateTagRequest{
		MoveUid: req.MoveUid,
		TagName: req.TagName,
		Remark:  req.Remark,
	}
	// 调用服务层创建标签
	tag, err := service.CreateTag(userUid.(string), serviceReq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建标签失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": "创建成功",
		"tag":      tag,
	})
}

// UpdateTagRequest 编辑标签请求参数
type UpdateTagRequest struct {
	TagUid     string `json:"tag_uid" binding:"required,uuid"`     // 标签UID
	TagName    string `json:"tag_name" binding:"required,max=100"` // 标签名称
	Remark     string `json:"remark" binding:"max=500"`            // 标签备注
	IsVerified int    `json:"is_verified" binding:"oneof=0 1"`     // 是否核销(0-未核销,1-已核销)
	Status     int    `json:"status" binding:"omitempty,oneof=0 1 2"` // 标签状态(0-正常,1-锁定,2-已完成)
}

// UpdateTag 编辑标签接口
// GetTagDetailRequest 获取标签详情请求参数
 type GetTagDetailRequest struct {
	TagUid string `json:"tag_uid" binding:"required,uuid"` // 标签UID
 }

 // GetTagDetail 获取标签详情接口
 func GetTagDetail(c *gin.Context) {
	var req GetTagDetailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户UID
	userUid, exists := c.Get("userUid")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    common.CodeUserNotLogin,
			"message": common.CodeMessage[common.CodeUserNotLogin],
		})
		return
	}

	// 调用服务层获取标签详情
	tag, err := service.GetTagDetail(userUid.(string), req.TagUid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取标签详情失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": "获取成功",
		"tag":  tag,
	})
}
func UpdateTag(c *gin.Context) {
	var req UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户UID
	userUid, exists := c.Get("userUid")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    common.CodeUserNotLogin,
			"message": common.CodeMessage[common.CodeUserNotLogin],
		})
		return
	}

	// 调用服务层更新标签
	tag, err := service.UpdateTag(userUid.(string), service.UpdateTagRequest{
		TagUid:     req.TagUid,
		TagName:    req.TagName,
		Remark:     req.Remark,
		IsVerified: req.IsVerified,
	})
	if err != nil {
		if err.Error() == "用户无此标签" {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeTagNotFound,
				"message": common.CodeMessage[common.CodeTagNotFound],
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新标签失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": "创建成功",
		"tag":      tag,
	})
}

// DeleteTagRequest 删除标签请求参数
type DeleteTagRequest struct {
	TagUid    string `json:"tag_uid" binding:"required,uuid"` // 标签UID
	IsDeleted int    `json:"is_deleted" binding:"oneof=0 1"`  // 是否删除(0-未删除,1-已删除)
}

// DeleteTag 删除标签接口
func DeleteTag(c *gin.Context) {
	var req DeleteTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户UID
	userUid, exists := c.Get("userUid")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    common.CodeUserNotLogin,
			"message": common.CodeMessage[common.CodeUserNotLogin],
		})
		return
	}

	// 调用服务层删除标签
	err := service.DeleteTag(userUid.(string), req.TagUid, req.IsDeleted)
	if err != nil {
		if err.Error() == "用户无此标签记录" {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeTagNotFound,
				"message": common.CodeMessage[common.CodeTagNotFound],
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "删除标签失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": "操作成功",
	})
}

// VerifyTagRequest 核销标签请求参数
type VerifyTagRequest struct {
	TagUid     string `json:"tag_uid" binding:"required,uuid"` // 标签UID
	IsVerified int    `json:"is_verified" binding:"oneof=0 1"` // 是否核销(0-未核销,1-已核销)
}

// VerifyTag 核销标签接口
func VerifyTag(c *gin.Context) {
	var req VerifyTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户UID
	userUid, exists := c.Get("userUid")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    common.CodeUserNotLogin,
			"message": common.CodeMessage[common.CodeUserNotLogin],
		})
		return
	}

	// 调用服务层核销标签
	err := service.VerifyTag(userUid.(string), req.TagUid, req.IsVerified)
	if err != nil {
		if err.Error() == "用户无此标签记录" {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeTagNotFound,
				"message": common.CodeMessage[common.CodeTagNotFound],
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "核销标签失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": "操作成功",
	})
}

// GetTagListRequest 标签列表请求参数
type GetTagListRequest struct {
	MoveUid  string `json:"move_uid" binding:"required,uuid"` // 搬运UID
	Page     int    `json:"page" binding:"min=1"`             // 页码
	PageSize int    `json:"page_size" binding:"min=1,max=50"` // 每页条数
}

// GetTagList 获取标签列表接口
func GetTagList(c *gin.Context) {
	var req GetTagListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户UID
	userUid, exists := c.Get("userUid")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    common.CodeUserNotLogin,
			"message": common.CodeMessage[common.CodeUserNotLogin],
		})
		return
	}

	// 调用服务层获取标签列表
	tags, total, err := service.GetTagList(userUid.(string), req.MoveUid, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取标签列表失败: " + err.Error(),
		})
		return
	}

	// 计算总页数
	totalPages := (total + int64(req.PageSize) - 1) / int64(req.PageSize)

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": common.CodeSuccess,
		"tags": tags,
		"pagination": gin.H{
			"total":      total,
			"page":       req.Page,
			"pageSize":   req.PageSize,
			"totalPages": totalPages,
		},
	})
}

// GeneratePDFRequest 生成PDF请求参数
type GeneratePDFRequest struct {
	MoveUid string `json:"move_uid" binding:"required,uuid"` // 搬运UID
}

// GeneratePDF 生成标签PDF接口
func GeneratePDF(c *gin.Context) {
	var req GeneratePDFRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户UID
	userUid, exists := c.Get("userUid")
	if !exists {
		c.JSON(http.StatusOK, gin.H{
			"code":    common.CodeUserNotLogin,
			"message": common.CodeMessage[common.CodeUserNotLogin],
		})
		return
	}

	// 调用服务层生成PDF
	pdfBytes, err := service.GenerateTagPDF(userUid.(string), req.MoveUid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "生成PDF失败: " + err.Error(),
		})
		return
	}

	// 设置响应头，返回PDF文件
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=tags.pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}
