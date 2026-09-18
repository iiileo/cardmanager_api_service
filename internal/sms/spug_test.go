package sms

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
)

func TestSpugSender_SendCode(t *testing.T) {
	var gotPath string
	var gotBody spugRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		defer r.Body.Close()
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_ = json.NewEncoder(w).Encode(spugResponse{
			Code: 200, Msg: "请求成功", RequestID: "test-req-1",
		})
	}))
	defer srv.Close()

	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	sender, err := NewSpugSender(SpugConfig{
		TemplateCode:    "tpl_abc",
		WithTTL:         true,
		HTTPClient:      srv.Client(),
		baseURLOverride: srv.URL,
	}, log)
	if err != nil {
		t.Fatal(err)
	}

	if err := sender.SendCode(context.Background(), SendCodeInput{
		Phone: "13800138000", Code: "654321", Scene: "login", TTLMinutes: 5,
	}); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/sms/tpl_abc" {
		t.Fatalf("path=%s", gotPath)
	}
	if gotBody.To != "13800138000" || gotBody.Code != "654321" || gotBody.Number != "5" {
		t.Fatalf("body=%+v", gotBody)
	}
}

func TestSpugSender_Rejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(spugResponse{Code: 400, Msg: "模板编码无效"})
	}))
	defer srv.Close()

	log := logger.NewLogger(&config.Config{Logging: config.LoggingConfig{Level: "error"}})
	sender, err := NewSpugSender(SpugConfig{
		TemplateCode:    "bad",
		HTTPClient:      srv.Client(),
		baseURLOverride: srv.URL,
	}, log)
	if err != nil {
		t.Fatal(err)
	}
	err = sender.SendCode(context.Background(), SendCodeInput{Phone: "13800138000", Code: "123456"})
	if err == nil {
		t.Fatal("expected error")
	}
}
