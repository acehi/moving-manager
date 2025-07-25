package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 所有模型的基础结构，包含公共字段
type BaseModel struct {
	CreatedAt int64 `gorm:"column:created_at" json:"created_at"`           // 创建时间戳
	UpdatedAt int64 `gorm:"column:updated_at" json:"updated_at"`           // 更新时间戳
	IsDeleted int   `gorm:"column:is_deleted;default:0" json:"is_deleted"` // 是否删除(0-未删除,1-已删除)
	DeletedAt int64 `gorm:"column:delete_at" json:"delete_at"`             // 删除时间戳
}

// BeforeCreate 创建前钩子，设置创建时间戳
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	b.CreatedAt = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子，设置更新时间戳
func (b *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	b.UpdatedAt = time.Now().Unix()
	return nil
}
