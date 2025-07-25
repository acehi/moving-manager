package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"gorm.io/gorm"

	"movingManager/model"
)

// UserAuth 用户登录/注册业务处理
// 如果用户存在则更新token，不存在则创建新用户
func UserAuth(mobile string) (*model.UserModel, string, error) {
	var user model.UserModel

	// 生成新的授权码(token)
	token, err := func() (string, error) {
		if user.ID == 0 {
			// 新用户需要先生成盐值
			salt, err := generateSalt()
			if err != nil {
				return "", fmt.Errorf("生成盐值失败: %v", err)
			}
			return generateAuthCode(mobile, salt)
		}
		// 老用户使用已有盐值
		return generateAuthCode(mobile, user.Salt)
	}()
	if err != nil {
		return nil, "", fmt.Errorf("生成授权码失败: %v", err)
	}

	// 查询用户是否存在
	if err := user.GetByMobile(mobile); err != nil {
		if err == gorm.ErrRecordNotFound {
			// 用户不存在，创建新用户
			salt, saltErr := generateSalt()
			if saltErr != nil {
				return nil, "", fmt.Errorf("生成盐值失败: %v", saltErr)
			}
			newUser := model.UserModel{
				Mobile:   mobile,
				UserName: getUserNameFromMobile(mobile),
				AuthCode: token,
				Salt:     salt,
			}

			if createErr := newUser.Create(); createErr != nil {
				return nil, "", fmt.Errorf("创建用户失败: %v", createErr)
			}
			return &newUser, token, nil
		}
		// 查询出错
		return nil, "", fmt.Errorf("查询用户失败: %v", err)
	} else {
		// 用户存在，更新token
		if err := user.UpdateAuthCode(token); err != nil {
			return nil, "", fmt.Errorf("更新授权码失败: %v", err)
		}
		return &user, token, nil
	}
}

// generateAuthCode 生成授权码
// 使用手机号+当前时间戳+用户盐值进行哈希生成唯一授权码
func generateAuthCode(mobile, salt string) (string, error) {
	// 获取当前时间戳(秒级)
	timestamp := time.Now().Unix()
	// 组合手机号、时间戳和盐值并哈希
	data := []byte(fmt.Sprintf("%s:%d:%s", mobile, timestamp, salt))
	hash := sha256.Sum256(data)
	return base64.URLEncoding.EncodeToString(hash[:]), nil
}

// generateSalt 生成用户唯一盐值
func generateSalt() (string, error) {
	saltBytes := make([]byte, 16)
	_, err := rand.Read(saltBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(saltBytes), nil
}

// getUserNameFromMobile 从手机号提取用户名(后四位)
func getUserNameFromMobile(mobile string) string {
	if len(mobile) >= 4 {
		return mobile[len(mobile)-4:]
	}
	return mobile
}
