package sms

import (
	"fmt"
	"strings"

	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
)

// NewSender 按 auth.sms_provider 创建发送通道。
//   - dev（默认且 sms_dev_mode=true）：不发真实短信
//   - spug：Spug 推送助手 https://push.spug.cc/guide/sms
func NewSender(cfg *config.Config, log *logger.Logger) (Sender, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Auth.SmsProvider))
	if provider == "" {
		if cfg.Auth.SmsDevMode {
			provider = "dev"
		} else {
			provider = "spug"
		}
	}

	switch provider {
	case "dev", "noop", "local":
		return NewDevSender(log), nil
	case "spug":
		return NewSpugSender(SpugConfig{
			TemplateCode: cfg.Auth.SmsSpugTemplateCode,
			WithTTL:      cfg.Auth.SmsSpugWithTTL,
		}, log)
	default:
		return nil, fmt.Errorf("unsupported auth.sms_provider: %q (want: dev|spug)", provider)
	}
}
