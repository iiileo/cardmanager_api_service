package service

import (
	"context"
	"strings"

	"card_manager/api_service/internal/api/dto"
	ierr "card_manager/api_service/internal/errors"
	"card_manager/api_service/internal/logger"
)

type HelloService interface {
	SayHello(ctx context.Context, req dto.HelloRequest) (*dto.HelloResponse, error)
}

type helloService struct {
	log *logger.Logger
}

func NewHelloService(log *logger.Logger) HelloService {
	return &helloService{log: log}
}

func (s *helloService) SayHello(ctx context.Context, req dto.HelloRequest) (*dto.HelloResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ierr.NewError("name is required").
			WithHint("Provide a non-empty name query or body field").
			Mark(ierr.ErrValidation)
	}

	s.log.Info(ctx, "hello requested", "name", name)
	return &dto.HelloResponse{
		Message: "Hello, " + name + "!",
	}, nil
}
