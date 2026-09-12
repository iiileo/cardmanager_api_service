package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Auth     AuthConfig     `mapstructure:"auth"`
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
	return nil
}
