package model

import (
	"time"
)

// APIResponse 统一API响应结构
type APIResponse struct {
	Code    int         `json:"code"`           // 状态码
	Message string      `json:"message"`        // 消息
	Data    interface{} `json:"data,omitempty"` // 数据
	Time    time.Time   `json:"time"`           // 响应时间（ISO8601格式）
}

// NewSuccessResponse 创建成功响应
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
		Time:    time.Now(),
	}
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *APIResponse {
	return &APIResponse{
		Code:    code,
		Message: message,
		Time:    time.Now(),
	}
}

// 预定义错误码
const (
	CodeSuccess            = 0     // 成功
	CodeInvalidParams      = 10001 // 参数错误
	CodeUnauthorized       = 10002 // 未授权
	CodeForbidden          = 10003 // 禁止访问
	CodeNotFound           = 10004 // 资源不存在
	CodeConflict           = 10005 // 资源冲突
	CodeInternalError      = 10006 // 内部错误
	CodeWechatAuthFailed   = 20001 // 微信授权失败
	CodePaymentFailed      = 30001 // 支付失败
	CodeOrderStatusError   = 40001 // 订单状态错误
	CodeDriverNotAvailable = 50001 // 司机不可用
)

// 预定义错误消息
const (
	MsgSuccess            = "success"
	MsgInvalidParams      = "invalid parameters"
	MsgUnauthorized       = "unauthorized"
	MsgForbidden          = "forbidden"
	MsgNotFound           = "not found"
	MsgConflict           = "resource conflict"
	MsgInternalError      = "internal server error"
	MsgWechatAuthFailed   = "wechat authentication failed"
	MsgPaymentFailed      = "payment failed"
	MsgOrderStatusError   = "order status error"
	MsgDriverNotAvailable = "driver not available"
)
