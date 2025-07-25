package database

import (
	"fmt"
	"os"

	"time"

	"gopkg.in/yaml.v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接实例
var DB *gorm.DB

// Config 数据库配置结构体
type Config struct {
	PostgreSQL struct {
		Host            string `yaml:"host"`
		Port            string `yaml:"port"`
		User            string `yaml:"user"`
		Password        string `yaml:"password"`
		Dbname          string `yaml:"dbname"`
		Sslmode         string `yaml:"sslmode"`
		MaxOpenConns    int    `yaml:"max_open_conns"`
		MaxIdleConns    int    `yaml:"max_idle_conns"`
		ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	} `yaml:"postgresql"`
}

// InitDB 初始化数据库连接
func InitDB() error {
	// 加载配置文件
	config, err := loadConfig("/Users/q/Q/code/products/movingManager/backend/config/database.yaml")
	if err != nil {
		return fmt.Errorf("加载数据库配置失败: %v", err)
	}

	// 构建DSN
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.PostgreSQL.Host,
		config.PostgreSQL.Port,
		config.PostgreSQL.User,
		config.PostgreSQL.Password,
		config.PostgreSQL.Dbname,
		config.PostgreSQL.Sslmode,
	)

	// 连接数据库
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 显示SQL日志
	})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %v", err)
	}

	// 设置连接池
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("获取数据库连接池失败: %v", err)
	}
	sqlDB.SetMaxOpenConns(config.PostgreSQL.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.PostgreSQL.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(config.PostgreSQL.ConnMaxLifetime) * time.Second)

	return err
}

// loadConfig 从YAML文件加载配置
func loadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
