package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"card_manager/api_service/internal/logger"
)

const spugBaseURL = "https://push.spug.cc"

// SpugConfig Spug 推送助手短信验证码（https://push.spug.cc/guide/sms）。
type SpugConfig struct {
	// TemplateCode 控制台模板编码，同时是调用凭证；出现在路径 /sms/<TEMPLATE_CODE>。
	TemplateCode string
	// WithTTL 为 true 时请求体带 number（有效分钟数），对应「带有效时长」的官方模板。
	WithTTL bool
	// HTTPClient 可选；默认 10s 超时。
	HTTPClient *http.Client
	// baseURLOverride 仅测试注入；生产固定为 spugBaseURL。
	baseURLOverride string
}

// SpugSender POST /sms/<TEMPLATE_CODE> 发送验证码。
type SpugSender struct {
	cfg SpugConfig
	log *logger.Logger
	hc  *http.Client
}

func NewSpugSender(cfg SpugConfig, log *logger.Logger) (*SpugSender, error) {
	code := strings.TrimSpace(cfg.TemplateCode)
	if code == "" {
		return nil, fmt.Errorf("sms.spug.template_code is required")
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.baseURLOverride), "/")
	if base == "" {
		base = spugBaseURL
	}
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &SpugSender{
		cfg: SpugConfig{
			TemplateCode:    code,
			WithTTL:         cfg.WithTTL,
			baseURLOverride: base,
		},
		log: log,
		hc:  hc,
	}, nil
}

type spugRequest struct {
	To     string `json:"to"`
	Code   string `json:"code"`
	Number string `json:"number,omitempty"`
}

type spugResponse struct {
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	RequestID string `json:"request_id"`
}

func (s *SpugSender) SendCode(ctx context.Context, in SendCodeInput) error {
	phone := strings.TrimSpace(in.Phone)
	code := strings.TrimSpace(in.Code)
	if phone == "" || code == "" {
		return fmt.Errorf("phone and code are required")
	}

	body := spugRequest{To: phone, Code: code}
	if s.cfg.WithTTL {
		mins := in.TTLMinutes
		if mins <= 0 {
			mins = 5
		}
		body.Number = fmt.Sprintf("%d", mins)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal spug sms body: %w", err)
	}

	url := s.cfg.baseURLOverride + "/sms/" + s.cfg.TemplateCode
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("new spug sms request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.hc.Do(req)
	if err != nil {
		return fmt.Errorf("spug sms request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read spug sms response: %w", err)
	}

	var out spugResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return fmt.Errorf("decode spug sms response (http=%d): %w; body=%s", resp.StatusCode, err, truncate(string(raw), 256))
	}
	// code=200 仅表示受理成功，投递结果需用 request_id 查询。
	if out.Code != 200 {
		msg := out.Msg
		if msg == "" {
			msg = fmt.Sprintf("http=%d body=%s", resp.StatusCode, truncate(string(raw), 256))
		}
		return fmt.Errorf("spug sms rejected: %s", msg)
	}

	s.log.Info(ctx, "spug sms accepted",
		"phone", phone,
		"scene", in.Scene,
		"request_id", out.RequestID,
	)
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
