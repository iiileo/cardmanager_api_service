package service

import (
	"context"
	"testing"

	"card_manager/api_service/internal/api/dto"
	"card_manager/api_service/internal/config"
	"card_manager/api_service/internal/logger"
)

func TestHelloService_SayHello(t *testing.T) {
	svc := NewHelloService(logger.NewLogger(&config.Config{
		Logging: config.LoggingConfig{Level: "error"},
	}))

	tests := []struct {
		name    string
		req     dto.HelloRequest
		want    string
		wantErr bool
	}{
		{name: "ok", req: dto.HelloRequest{Name: "world"}, want: "Hello, world!"},
		{name: "trim", req: dto.HelloRequest{Name: "  Dora  "}, want: "Hello, Dora!"},
		{name: "empty", req: dto.HelloRequest{Name: ""}, wantErr: true},
		{name: "blank", req: dto.HelloRequest{Name: "   "}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.SayHello(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.Message != tt.want {
				t.Fatalf("got=%q want=%q", got.Message, tt.want)
			}
		})
	}
}
