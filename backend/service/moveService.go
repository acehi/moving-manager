package service

import (
	"fmt"

	"gorm.io/gorm"

	"movingManager/model"
)

// 定义业务常量
const (
	DefaultTagCount     = 0 // 默认标签数量
	DefaultUnverified   = 0 // 默认未核销数量
	DefaultNotCompleted = 0 // 默认未完成状态
	DefaultNotDeleted   = 0 // 默认未删除状态
)

// CreateMove 创建搬运业务处理
func CreateMove(userUid string, moveTime int64, startLocation, endLocation, remark string) (*model.MoveModel, error) {
	// 创建搬运记录
	move := model.MoveModel{
		UserUid:             userUid,
		MoveAt:              moveTime,
		StartLocation:       startLocation,
		EndLocation:         endLocation,
		Remark:              remark,
		TagCount:            DefaultTagCount,
		UnverifiedTagCount:  DefaultUnverified,
		VerifiedTagCount:    DefaultUnverified,
		IsCompleted:         DefaultNotCompleted,
	}

	if err := move.Create(); err != nil {
		return nil, fmt.Errorf("创建搬运记录失败: %v", err)
	}

	return &move, nil
}

// GetMoveDetail 获取搬运详情业务处理
func GetMoveDetail(userUid, moveUid string, onlyUndeleted bool) (*model.MoveModel, error) {
	var move model.MoveModel

	// 调用model层查询方法，根据参数决定是否包含已删除记录
	if err := move.GetByUID(userUid, moveUid, onlyUndeleted); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户无此搬运记录")
		}
		return nil, fmt.Errorf("查询搬运记录失败: %v", err)
	}

	return &move, nil
}

// UpdateMoveRequest 更新搬运请求参数
type UpdateMoveRequest struct {
	MoveUid       string `json:"move_uid"`       // 搬运UID
	MoveAt        int64  `json:"move_at"`        // 搬运时间戳
	StartLocation string `json:"start_location"` // 出发地
	EndLocation   string `json:"end_location"`   // 目的地
	Remark        string `json:"remark"`         // 备注
	IsCompleted   int    `json:"is_completed"`   // 是否完成(0-未完成,1-已完成)
}

func UpdateMove(userUid string, req UpdateMoveRequest) (*model.MoveModel, error) {
	var move model.MoveModel

	// 调用model层查询方法
	if err := move.GetByUID(userUid, req.MoveUid, true); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户无此搬运记录")
		}
		return nil, fmt.Errorf("查询搬运记录失败: %v", err)
	}

	// 更新字段
	move.MoveAt = req.MoveAt
	move.StartLocation = req.StartLocation
	move.EndLocation = req.EndLocation
	move.Remark = req.Remark
	move.IsCompleted = req.IsCompleted

	// 调用model层更新方法
	if err := move.Update(); err != nil {
		return nil, fmt.Errorf("更新搬运记录失败: %v", err)
	}

	return &move, nil
}

// DeleteMove 删除搬运业务处理
func DeleteMove(userUid, moveUid string, isDeleted int) error {
	var move model.MoveModel

	// 恢复操作时需要查找已删除记录（isDeleted=0表示恢复）
	onlyUndeleted := isDeleted != 0
	// 调用model层查询方法
	if err := move.GetByUID(userUid, moveUid, onlyUndeleted); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("用户无此搬运记录")
		}
		return fmt.Errorf("查询搬运记录失败: %v", err)
	}

	// 调用model层更新删除状态方法
	return move.UpdateDeleteStatus(isDeleted)
}

// GetMoveList 获取搬运列表业务处理
func GetMoveList(userUid string, page, pageSize int, onlyUndeleted bool) ([]model.MoveModel, int64, error) {
	var move model.MoveModel
	moves, total, err := move.ListByUser(userUid, page, pageSize, onlyUndeleted)
	if err != nil {
		return nil, 0, fmt.Errorf("查询搬运列表失败: %v", err)
	}
	return moves, total, nil
}
