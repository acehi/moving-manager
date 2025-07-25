package migrate

import (
	"fmt"
	"movingManager/database"
	"movingManager/model"
)

// Migrate 执行数据库迁移
func Migrate() {
	// 自动迁移数据表结构
	db := database.DB
	err := db.AutoMigrate(
		&model.UserModel{},
		&model.MoveModel{},
		&model.TagModel{},
	)
	if err != nil {
		// 处理迁移错误
		fmt.Printf("自动迁移失败: %v\n", err)
		return
	}

	// 设置ID自增起始值为10000（PostgreSQL）
	// 先检查序列是否存在，不存在则创建
	sequences := []string{"users_id_seq", "moves_id_seq", "tags_id_seq"}
	for _, seq := range sequences {
		// 检查序列是否存在
		var exists bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_class WHERE relname = ?)", seq).Scan(&exists).Error
		if err != nil {
			fmt.Printf("检查序列 %s 失败: %v\n", seq, err)
			continue
		}
		
		if exists {
			// 重置序列
			err := db.Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 10000", seq)).Error
			if err != nil {
				fmt.Printf("重置序列 %s 失败: %v\n", seq, err)
			}
		} else {
			fmt.Printf("序列 %s 不存在，跳过重置\n", seq)
		}
	}
}
