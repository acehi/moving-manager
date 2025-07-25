package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"movingManager/database"
)

// UserModel 用户表模型
// 存储用户基本信息，包括登录凭证和状态

type UserModel struct {
	ID        uint   `gorm:"primarykey" json:"id"`                                // 主键ID
	UserUid   string `gorm:"column:user_uid;uniqueIndex;size:36" json:"user_uid"` // 用户唯一标识
	Mobile      string `gorm:"column:mobile;uniqueIndex;size:20" json:"mobile"`     // 手机号(登录账号)
	UserName    string `gorm:"column:user_name;size:50" json:"user_name"`           // 用户名
	AuthCode  string `gorm:"column:authorization_code;size:100" json:"-"`         // 登录授权码(token)
	Salt      string `gorm:"column:salt;size:50" json:"-"`                        // 用户盐值
	BaseModel        // 嵌入基础模型
}

// TableName 设置表名
func (u *UserModel) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子：生成UUID作为用户唯一标识
func (u *UserModel) BeforeCreate(tx *gorm.DB) error {
	if u.UserUid == "" {
		u.UserUid = uuid.New().String()
	}
	return nil
}

// Create 创建用户记录
func (u *UserModel) Create() error {
	db := database.DB
	return db.Create(u).Error
}

// GetByMobile 根据手机号查询有效用户
func (u *UserModel) GetByMobile(mobile string) error {
	db := database.DB
	return db.Where("mobile = ? AND is_deleted = 0", mobile).First(u).Error
}

// UpdateAuthCode 更新用户授权码
func (u *UserModel) UpdateAuthCode(newCode string) error {
	db := database.DB
	return db.Model(u).Update("authorization_code", newCode).Error
}
