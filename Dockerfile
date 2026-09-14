# 多阶段构建：编译 Go 服务，最终镜像只保留二进制与配置。
FROM golang:1.25-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /out/server ./server
# Docker 默认配置；本地开发仍用 internal/config/config.yaml
COPY internal/config/config.docker.yaml ./internal/config/config.yaml

EXPOSE 8080
CMD ["./server"]
