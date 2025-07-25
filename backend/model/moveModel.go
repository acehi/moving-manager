package model

import (
	// 由于存在导入循环问题，将数据库依赖通过接口或依赖注入方式处理，此处暂时移除该导入
	// 实际使用时需要通过依赖注入传递数据库连接
	// 由于存在导入循环问题，移除该导入，实际使用时通过依赖注入传递数据库连接
	"movingManager/database"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MoveModel 搬运记录表模型
// 存储用户的搬运任务基本信息及标签统计数据
type MoveModel struct {
	ID                 uint   `gorm:"primarykey;autoIncrement" json:"id"`                                              // 主键ID
	MoveUid            string `gorm:"column:move_uid;uniqueIndex;size:36" json:"move_uid"`               // 搬运唯一标识
	UserUid            string `gorm:"column:user_uid;index;size:36" json:"user_uid"`                     // 所属用户UID
	MoveAt             int64  `gorm:"column:move_at;type:bigint" json:"move_at"`                          // 搬运时间戳（Unix时间）
	StartLocation      string `gorm:"column:start_location;size:100" json:"start_location"`              // 出发地
	EndLocation        string `gorm:"column:end_location;size:100" json:"end_location"`                  // 目的地
	TagCount           int    `gorm:"column:tag_count;default:0" json:"tag_count"`                       // 标签总数
	VerifiedTagCount   int    `gorm:"column:verified_tag_count;default:0" json:"verified_tag_count"`     // 已核销标签数
	UnverifiedTagCount int    `gorm:"column:unverified_tag_count;default:0" json:"unverified_tag_count"` // 未核销标签数
	IsCompleted        int    `gorm:"column:is_completed;default:0" json:"is_completed"`                 // 是否完成(0-未完成,1-已完成)
	Remark             string `gorm:"column:remark;size:500" json:"remark"`                              // 备注信息
	BaseModel                 // 嵌入基础模型
}

// TableName 设置表名
func (m *MoveModel) TableName() string {
	return "moves"
}

// BeforeCreate 创建前钩子：生成UUID作为搬运唯一标识
func (m *MoveModel) BeforeCreate(tx *gorm.DB) error {
	if m.MoveUid == "" {
		m.MoveUid = uuid.New().String()
	}
	return nil
}

// Create 插入搬运记录到数据库
func (m *MoveModel) Create() error {
	return database.DB.Create(m).Error
}

// GetByUID 根据用户UID和搬运UID查询记录
// onlyUndeleted 控制是否只查询未删除记录
func (m *MoveModel) GetByUID(userUid, moveUid string, onlyUndeleted bool) error {
	where := "user_uid = ? AND move_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}
	return database.DB.Where(where, userUid, moveUid).First(m).Error
}

// Update 更新搬运记录
func (m *MoveModel) Update() error {
	return database.DB.Save(m).Error
}

// UpdateDeleteStatus 更新删除状态
func (m *MoveModel) UpdateDeleteStatus(isDeleted int) error {
	m.IsDeleted = isDeleted
	if isDeleted == 1 {
		m.DeletedAt = time.Now().Unix()
	}

	return m.Update()
}

// ListByUser 获取用户搬运列表
// onlyUndeleted 控制是否只查询未删除记录
func (m *MoveModel) ListByUser(userUid string, page, pageSize int, onlyUndeleted bool) ([]MoveModel, int64, error) {
	var moves []MoveModel
	var total int64
	offset := (page - 1) * pageSize
	// 使用全局数据库连接替代未初始化的局部变量
	db := database.DB.Model(&MoveModel{})

	where := "user_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}

	// 查询总数
	if err := db.Where(where, userUid).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	// 排序规则：未完成的搬运在前（is_completed=0），然后按搬运时间正序排列
	if err := db.Where(where, userUid).Order("is_completed ASC, move_at ASC").Limit(pageSize).Offset(offset).Find(&moves).Error; err != nil {
		return nil, 0, err
	}

	return moves, total, nil
}
