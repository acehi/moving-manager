package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"movingManager/database"
)

// TagModel 标签表模型
// 存储搬运任务下的标签信息，包含标签状态和关联关系
type TagModel struct {
	ID         uint   `gorm:"primarykey;autoIncrement" json:"id"`                              // 主键ID
	TagUid     string `gorm:"column:tag_uid;uniqueIndex;size:36" json:"tag_uid"` // 标签唯一标识
	UserUid    string `gorm:"column:user_uid;index;size:36" json:"user_uid"`     // 所属用户UID
	MoveUid    string `gorm:"column:move_uid;index;size:36" json:"move_uid"`     // 所属搬运UID
	TagName    string `gorm:"column:tag_name;size:100" json:"tag_name"`          // 标签名称
	Remark     string `gorm:"column:remark;size:500" json:"remark"`              // 标签备注
	IsVerified int    `gorm:"column:is_verified;default:0" json:"is_verified"`   // 是否核销(0-未核销,1-已核销)
	BaseModel         // 嵌入基础模型
}

// TableName 设置表名
func (t *TagModel) TableName() string {
	return "tags"
}

// BeforeCreate 创建前钩子：生成UUID作为标签唯一标识
func (t *TagModel) BeforeCreate(tx *gorm.DB) error {
	if t.TagUid == "" {
		t.TagUid = uuid.New().String()
	}
	return nil
}

// Create 插入标签记录到数据库
func (t *TagModel) Create() error {
	return database.DB.Create(t).Error
}

// CreateTx 事务中插入标签记录
func (t *TagModel) CreateTx(tx *gorm.DB) error {
	return tx.Create(t).Error
}

// GetByUID 根据用户UID和标签UID查询记录
// onlyUndeleted 控制是否只查询未删除记录
func (t *TagModel) GetByUID(userUid, tagUid string, onlyUndeleted bool) error {
	where := "user_uid = ? AND tag_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}
	return database.DB.Where(where, userUid, tagUid).First(t).Error
}

// GetByUserAndTagUidTx 事务中根据用户UID和标签UID查询记录
// onlyUndeleted 控制是否只查询未删除记录
func (t *TagModel) GetByUserAndTagUidTx(tx *gorm.DB, userUid, tagUid string, onlyUndeleted bool) error {
	where := "user_uid = ? AND tag_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}
	return tx.Where(where, userUid, tagUid).First(t).Error
}

// Update 更新标签记录
func (t *TagModel) Update() error {
	return database.DB.Save(t).Error
}

// UpdateTx 事务中更新标签记录
func (t *TagModel) UpdateTx(tx *gorm.DB) error {
	return tx.Save(t).Error
}

// UpdateDeleteStatus 更新删除状态
func (t *TagModel) UpdateDeleteStatus(isDeleted int) error {
	t.IsDeleted = isDeleted
	if isDeleted == 1 {
		t.DeletedAt = time.Now().Unix()
	}

	return t.Update()
}

// UpdateDeleteStatusTx 事务中更新删除状态
func (t *TagModel) UpdateDeleteStatusTx(tx *gorm.DB, isDeleted int) error {
	t.IsDeleted = isDeleted
	if isDeleted == 1 {
		t.DeletedAt = time.Now().Unix()
	}

	return t.UpdateTx(tx)
}

// GetMoveByUserAndMoveUidTx 事务中验证搬运记录是否存在且属于当前用户
// onlyUndeleted 控制是否只查询未删除记录
func (t *TagModel) GetMoveByUserAndMoveUidTx(tx *gorm.DB, userUid, moveUid string, onlyUndeleted bool) (*MoveModel, error) {
	var move MoveModel
	where := "user_uid = ? AND move_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}
	if err := tx.Where(where, userUid, moveUid).First(&move).Error; err != nil {
		return nil, err
	}
	return &move, nil
}

// GetMoveByMoveUidTx 事务中根据搬运UID查询搬运记录
// onlyUndeleted 控制是否只查询未删除记录
func (t *TagModel) GetMoveByMoveUidTx(tx *gorm.DB, moveUid string, onlyUndeleted bool) (*MoveModel, error) {
	var move MoveModel
	where := "move_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}
	if err := tx.Where(where, moveUid).First(&move).Error; err != nil {
		return nil, err
	}
	return &move, nil
}

// UpdateMoveTagCountTx 事务中更新搬运记录的标签统计
func (t *TagModel) UpdateMoveTagCountTx(tx *gorm.DB, move *MoveModel, tagCount, verifiedTagCount, unverifiedTagCount int) error {
	updates := make(map[string]interface{})
	if tagCount != 0 {
		updates["tag_count"] = gorm.Expr("tag_count + ?", tagCount)
	}
	if verifiedTagCount != 0 {
		updates["verified_tag_count"] = gorm.Expr("verified_tag_count + ?", verifiedTagCount)
	}
	if unverifiedTagCount != 0 {
		updates["unverified_tag_count"] = gorm.Expr("unverified_tag_count + ?", unverifiedTagCount)
	}
	return tx.Model(move).Updates(updates).Error
}

// VerifyTagTx 事务中核销标签
func (t *TagModel) VerifyTagTx(tx *gorm.DB, userUid, tagUid string, isVerified int) error {
	// 查询标签并验证所有权（恢复操作需要包含已删除记录）
	var tag TagModel
	if err := tx.Where("user_uid = ? AND tag_uid = ?", userUid, tagUid).First(&tag).Error; err != nil {
		return err
	}

	// 如果状态没变，直接返回
	if tag.IsVerified == isVerified {
		return nil
	}

	// 查询关联的搬运记录
	var move MoveModel
	if err := tx.Where("move_uid = ?", tag.MoveUid).First(&move).Error; err != nil {
		return err
	}

	// 更新搬运记录的标签统计
	var verifiedDelta, unverifiedDelta int
	if isVerified == 1 {
		verifiedDelta = 1
		unverifiedDelta = -1
	} else {
		verifiedDelta = -1
		unverifiedDelta = 1
	}

	if err := tx.Model(&move).Updates(map[string]interface{}{
		"verified_tag_count":   gorm.Expr("verified_tag_count + ?", verifiedDelta),
		"unverified_tag_count": gorm.Expr("unverified_tag_count + ?", unverifiedDelta),
	}).Error; err != nil {
		return err
	}

	// 更新标签核销状态
	tag.IsVerified = isVerified
	return tx.Save(&tag).Error
}

// GetTagDetail 获取标签详情
// onlyUndeleted 控制是否只查询未删除记录
func (t *TagModel) GetTagDetail(userUid, tagUid string, onlyUndeleted bool) (*TagModel, error) {
	var tag TagModel
	where := "user_uid = ? AND tag_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}
	if err := database.DB.Where(where, userUid, tagUid).First(&tag).Error; err != nil {
		return nil, err
	}
	return &tag, nil
}

// GetTagsByMove 获取搬运下的所有标签
// onlyUndeleted 控制是否只查询未删除记录
func (t *TagModel) GetTagsByMove(userUid, moveUid string, onlyUndeleted bool) ([]TagModel, error) {
	var tags []TagModel
	where := "user_uid = ? AND move_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}
	if err := database.DB.Where(where, userUid, moveUid).Order("created_at asc").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// ListByMove 获取搬运下的标签列表
// onlyUndeleted 控制是否只查询未删除记录
func (t *TagModel) ListByMove(userUid, moveUid string, page, pageSize int, onlyUndeleted bool) ([]TagModel, int64, error) {
	var tags []TagModel
	var total int64
	offset := (page - 1) * pageSize
	db := database.DB.Model(&TagModel{})

	where := "user_uid = ? AND move_uid = ?"
	if onlyUndeleted {
		where += " AND is_deleted = 0"
	}

	// 查询总数
	if err := db.Where(where, userUid, moveUid).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询列表
	if err := db.Where(where, userUid, moveUid).Order("is_verified asc").Order("created_at asc").Limit(pageSize).Offset(offset).Find(&tags).Error; err != nil {
		return nil, 0, err
	}

	return tags, total, nil
}
