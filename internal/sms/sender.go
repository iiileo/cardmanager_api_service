package sms

import "context"

// SendCodeInput 发送验证码入参（与具体通道无关）。
type SendCodeInput struct {
	Phone      string
	Code       string
	Scene      string // login | bind_phone 等，供日志/扩展
	TTLMinutes int    // 有效分钟数；部分模板（如 Spug 带有效期模板）必填
}

// Sender 短信验证码发送通道。新增厂商时实现本接口并在 NewSender 中注册即可。
type Sender interface {
	SendCode(ctx context.Context, in SendCodeInput) error
}
