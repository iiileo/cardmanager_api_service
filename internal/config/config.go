package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Postgres  PostgresConfig  `mapstructure:"postgres"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Auth      AuthConfig      `mapstructure:"auth"`
	AccessLog AccessLogConfig `mapstructure:"access_log"`
}

type AccessLogConfig struct {
	Enabled      bool     `mapstructure:"enabled"`
	FilePath     string   `mapstructure:"file_path"`
	MaxMemory    int      `mapstructure:"max_memory"`
	MaxBodyBytes int      `mapstructure:"max_body_bytes"`
	UIEnabled    bool     `mapstructure:"ui_enabled"`
	UIPath       string   `mapstructure:"ui_path"`
	SkipPaths    []string `mapstructure:"skip_paths"`
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
	Mode    string `mapstructure:"mode"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

type AuthConfig struct {
	JWTSecret         string `mapstructure:"jwt_secret"`
	AccessTTLSeconds  int64  `mapstructure:"access_ttl_seconds"`
	RefreshTTLSeconds int64  `mapstructure:"refresh_ttl_seconds"`
	SmsCodeTTLSeconds int64  `mapstructure:"sms_code_ttl_seconds"`
	SmsDevCode        string `mapstructure:"sms_dev_code"`
	SmsDevMode        bool   `mapstructure:"sms_dev_mode"`
	// SmsProvider 发送通道：dev | spug；空则 sms_dev_mode=true 用 dev，否则 spug。
	SmsProvider string `mapstructure:"sms_provider"`
	// Spug 推送助手短信（https://push.spug.cc/guide/sms）
	SmsSpugTemplateCode string `mapstructure:"sms_spug_template_code"`
	// SmsSpugWithTTL 为 true 时请求体带 number（分钟），对应带有效时长的官方模板。
	SmsSpugWithTTL bool `mapstructure:"sms_spug_with_ttl"`
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("internal/config")
	v.AddConfigPath(".")

	v.SetEnvPrefix("CARDMANAGER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server.address is required")
	}
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required")
	}
	if c.Auth.AccessTTLSeconds <= 0 {
		c.Auth.AccessTTLSeconds = 7200
	}
	if c.Auth.RefreshTTLSeconds <= 0 {
		c.Auth.RefreshTTLSeconds = 2592000
	}
	if c.Auth.SmsCodeTTLSeconds <= 0 {
		c.Auth.SmsCodeTTLSeconds = 300
	}
	if c.Auth.SmsDevCode == "" {
		c.Auth.SmsDevCode = "123456"
	}
	provider := strings.ToLower(strings.TrimSpace(c.Auth.SmsProvider))
	if provider == "" {
		if c.Auth.SmsDevMode {
			provider = "dev"
		} else {
			provider = "spug"
		}
	}
	c.Auth.SmsProvider = provider
	switch provider {
	case "dev", "noop", "local":
		// ok
	case "spug":
		if strings.TrimSpace(c.Auth.SmsSpugTemplateCode) == "" {
			return fmt.Errorf("auth.sms_spug_template_code is required when sms_provider=spug")
		}
	default:
		return fmt.Errorf("auth.sms_provider must be dev or spug, got %q", provider)
	}
	return nil
}
