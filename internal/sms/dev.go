package sms

import (
	"context"

	"card_manager/api_service/internal/logger"
)

// DevSender 本地/联调：不真正发短信，仅打日志。验证码由 Auth 在 sms_dev_mode 下回传 DevCode。
type DevSender struct {
	log *logger.Logger
}

func NewDevSender(log *logger.Logger) *DevSender {
	return &DevSender{log: log}
}

func (s *DevSender) SendCode(ctx context.Context, in SendCodeInput) error {
	s.log.Info(ctx, "sms skipped (dev sender)",
		"phone", in.Phone,
		"scene", in.Scene,
		"ttl_minutes", in.TTLMinutes,
	)
	return nil
}
