package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"movingManager/common"
	"movingManager/model"
	"movingManager/service"
)

// CreateMoveRequest 类型使用dto包中的定义
// CreateMoveRequest 创建搬运请求参数
type CreateMoveRequest struct {
	MoveAt        int64 `json:"move_at" binding:"required"` // 搬运时间戳(Unix时间)
	StartLocation string `json:"start_location" binding:"required,max=100"`      // 出发地
	EndLocation   string `json:"end_location" binding:"required,max=100"`        // 目的地
	Remark        string `json:"remark" binding:"max=500"`                       // 备注
}

// CreateMove 创建搬运接口
func CreateMove(c *gin.Context) {
	var req CreateMoveRequest
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

	// 提取参数并赋值给变量
	moveTime := req.MoveAt
	startLocation := req.StartLocation
	endLocation := req.EndLocation
	remark := req.Remark

	// 调用服务层创建搬运
	move, err := service.CreateMove(userUid.(string), moveTime, startLocation, endLocation, remark)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "创建搬运失败: " + err.Error(),
		})
		return
	}

	// 组装响应格式
	response := gin.H{
		"move_uid":             move.MoveUid,
		"move_at":              time.Unix(move.MoveAt, 0).Format("2006-01-02 15:04:05"),
		"start_location":       move.StartLocation,
		"end_location":         move.EndLocation,
		"tag_count":            move.TagCount,
		"verified_tag_count":   move.VerifiedTagCount,
		"unverified_tag_count": move.UnverifiedTagCount,
		"is_completed":         move.IsCompleted,
		"is_deleted":           move.IsDeleted,
		"remark":               move.Remark,
		"created_at":           time.Unix(move.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	if move.UpdatedAt > 0 {
		response["updated_at"] = time.Unix(move.UpdatedAt, 0).Format("2006-01-02 15:04:05")
	}

	if move.IsDeleted == 1 {
		response["delete_at"] = move.DeletedAt
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": "创建成功",
		"move":    response,
	})
}

// GetMoveDetailRequest 类型使用dto包中的定义
// GetMoveDetailRequest 搬运详情请求参数
type GetMoveDetailRequest struct {
	MoveUid string `json:"move_uid" binding:"required,uuid"` // 搬运UID
}

// GetMoveDetail 搬运详情接口
func GetMoveDetail(c *gin.Context) {
	var req GetMoveDetailRequest
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

	// 提取参数并赋值给变量
	moveUid := req.MoveUid

	// 调用服务层获取搬运详情
	modelMove, err := service.GetMoveDetail(userUid.(string), moveUid, true)
	if err != nil {
		if err.Error() == "用户无此搬运记录" {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeMoveNotFound,
				"message": common.CodeMessage[common.CodeMoveNotFound],
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取搬运详情失败: " + err.Error(),
		})
		return
	}

	// 组装响应格式
	response := gin.H{
		"move_uid":             modelMove.MoveUid,
		"move_at":              time.Unix(modelMove.MoveAt, 0).Format("2006-01-02 15:04:05"),
		"start_location":       modelMove.StartLocation,
		"end_location":         modelMove.EndLocation,
		"tag_count":            modelMove.TagCount,
		"verified_tag_count":   modelMove.VerifiedTagCount,
		"unverified_tag_count": modelMove.UnverifiedTagCount,
		"is_completed":         modelMove.IsCompleted,
		"is_deleted":           modelMove.IsDeleted,
		"remark":               modelMove.Remark,
		"created_at":           time.Unix(modelMove.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	if modelMove.UpdatedAt > 0 {
		response["updated_at"] = time.Unix(modelMove.UpdatedAt, 0).Format("2006-01-02 15:04:05")
	}

	if modelMove.IsDeleted == 1 {
		response["delete_time"] = modelMove.DeletedAt
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": common.CodeSuccess,
		"move": response,
	})
}

// UpdateMoveRequest 类型使用dto包中的定义
// UpdateMoveRequest 编辑搬运请求参数
type UpdateMoveRequest struct {
	MoveUid       string `json:"move_uid" binding:"required,uuid"`               // 搬运UID
	MoveAt        int64 `json:"move_at" binding:"required"` // 搬运时间戳(Unix时间)
	StartLocation string `json:"start_location" binding:"required,max=100"`      // 出发地
	EndLocation   string `json:"end_location" binding:"required,max=100"`        // 目的地
	Remark        string `json:"remark" binding:"max=500"`                       // 备注
	IsCompleted   int    `json:"is_completed" binding:"oneof=0 1"`               // 是否完成(0-未完成,1-已完成)
}

// UpdateMove 编辑搬运接口
func UpdateMove(c *gin.Context) {
	var req UpdateMoveRequest
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

	// 提取参数并赋值给变量
	moveUid := req.MoveUid
	startLocation := req.StartLocation
	endLocation := req.EndLocation
	remark := req.Remark
	isCompleted := req.IsCompleted

	moveTime := req.MoveAt

	// 封装服务层请求参数
	updateReq := service.UpdateMoveRequest{
		MoveUid:       moveUid,
		MoveAt:        moveTime,
		StartLocation: startLocation,
		EndLocation:   endLocation,
		Remark:        remark,
		IsCompleted:   isCompleted,
	}
	// 调用服务层更新搬运
	updatedMove, err := service.UpdateMove(userUid.(string), updateReq)
	if err != nil {
		if err.Error() == "用户无此搬运记录" {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeMoveNotFound,
				"message": common.CodeMessage[common.CodeMoveNotFound],
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "更新搬运失败: " + err.Error(),
		})
		return
	}

	// 组装响应格式
	response := gin.H{
		"move_uid":             updatedMove.MoveUid,
		"move_at":              updatedMove.MoveAt,
		"start_location":       updatedMove.StartLocation,
		"end_location":         updatedMove.EndLocation,
		"tag_count":            updatedMove.TagCount,
		"verified_tag_count":   updatedMove.VerifiedTagCount,
		"unverified_tag_count": updatedMove.UnverifiedTagCount,
		"is_completed":         updatedMove.IsCompleted,
		"is_deleted":           updatedMove.IsDeleted,
		"remark":               updatedMove.Remark,
		"created_at":           time.Unix(updatedMove.CreatedAt, 0).Format("2006-01-02 15:04:05"),
	}

	if updatedMove.UpdatedAt > 0 {
		response["updated_at"] = time.Unix(updatedMove.UpdatedAt, 0).Format("2006-01-02 15:04:05")
	}

	if updatedMove.IsDeleted == 1 {
		response["delete_time"] = updatedMove.DeletedAt
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code": common.CodeSuccess,
		"move": response,
	})
}

// DeleteMoveRequest 类型使用dto包中的定义
// DeleteMoveRequest 删除搬运请求参数
type DeleteMoveRequest struct {
	MoveUid   string `json:"move_uid" binding:"required,uuid"` // 搬运UID
	IsDeleted int    `json:"is_deleted" binding:"oneof=0 1"`   // 是否删除(0-未删除,1-已删除)
}

// DeleteMove 删除搬运接口
func DeleteMove(c *gin.Context) {
	var req DeleteMoveRequest
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

	// 提取参数并赋值给变量
	moveUid := req.MoveUid
	isDeleted := req.IsDeleted

	// 调用服务层删除搬运
	err := service.DeleteMove(userUid.(string), moveUid, isDeleted)
	if err != nil {
		if err.Error() == "用户无此搬运记录" {
			c.JSON(http.StatusOK, gin.H{
				"code":    common.CodeMoveNotFound,
				"message": common.CodeMessage[common.CodeMoveNotFound],
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "删除搬运失败: " + err.Error(),
		})
		return
	}

	// 组装响应格式
	response := gin.H{
		"move_uid":    moveUid,
		"is_deleted":  isDeleted,
		"delete_time": time.Now().Unix(),
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":    common.CodeSuccess,
		"message": "操作成功",
		"result":  response,
	})
}

// GetMoveListRequest 类型使用dto包中的定义
// GetMoveListRequest 搬运列表请求参数
type GetMoveListRequest struct {
	Page     int `json:"page" binding:"min=1"`             // 页码
	PageSize int `json:"page_size" binding:"min=1,max=50"` // 每页条数
}

// GetMoveList 搬运列表接口
func GetMoveList(c *gin.Context) {
	var req GetMoveListRequest
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

	// 提取参数并赋值给变量
	page := req.Page
	pageSize := req.PageSize

	// 调用服务层获取搬运列表
	moves, total, err := service.GetMoveList(userUid.(string), page, pageSize, true)
	if moves == nil {
		moves = []model.MoveModel{}
	}
	// 删除未使用的变量声明
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "获取搬运列表失败: " + err.Error(),
		})
		return
	}

	// 组装响应格式
	var responseList []gin.H
	for _, move := range moves {
		item := gin.H{
			"move_uid":             move.MoveUid,
			"move_at":              time.Unix(move.MoveAt, 0).Format("2006-01-02 15:04:05"),
			"start_location":       move.StartLocation,
			"end_location":         move.EndLocation,
			"tag_count":            move.TagCount,
			"verified_tag_count":   move.VerifiedTagCount,
			"unverified_tag_count": move.UnverifiedTagCount,
			"is_completed":         move.IsCompleted,
			"is_deleted":           move.IsDeleted,
			"remark":               move.Remark,
			"created_at":           time.Unix(move.CreatedAt, 0).Format("2006-01-02 15:04:05"),
		}

		if move.UpdatedAt > 0 {
			item["updated_at"] = time.Unix(move.UpdatedAt, 0).Format("2006-01-02 15:04:05")
		}

		if move.IsDeleted == 1 {
			item["delete_time"] = move.DeletedAt
		}

		responseList = append(responseList, item)
	}

	// 计算总页数
	totalPages := (total + int64(pageSize) - 1) / int64(pageSize)

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"code":     common.CodeSuccess,
		"message":  "获取成功",
		"moveList": responseList,
		"pagination": gin.H{
			"total":      total,
			"page":       page,
			"pageSize":   pageSize,
			"totalPages": totalPages,
		},
	})
}
