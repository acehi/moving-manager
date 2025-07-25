package common

// 响应码定义
const (
	CodeSuccess               = 200  // 成功
	CodeUserNotLogin          = 10000 // 用户未登录
	CodeUserNotRegistered     = 10001 // 用户未注册
	CodeMoveNotFound          = 20000 // 用户无此搬运记录
	CodeTagNotFound           = 30000 // 用户无此标签记录
)

// 响应消息映射
var CodeMessage = map[int]string{
	CodeSuccess:               "操作成功",
	CodeUserNotLogin:          "用户未登录",
	CodeUserNotRegistered:     "用户未注册",
	CodeMoveNotFound:          "用户无此搬运记录",
	CodeTagNotFound:           "用户无此标签记录",
}