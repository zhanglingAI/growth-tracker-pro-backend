package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	AI       AIConfig       `mapstructure:"ai"`
	WeChat   WeChatConfig   `mapstructure:"wechat"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Charset  string `mapstructure:"charset"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireTime int    `mapstructure:"expire_time"` // 秒
}

// AIConfig AI服务配置
type AIConfig struct {
	Provider   string `mapstructure:"provider"`   // deepseek, openai
	APIKey     string `mapstructure:"api_key"`
	Model      string `mapstructure:"model"`
	BaseURL    string `mapstructure:"base_url"`
	MaxTokens  int    `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// WeChatConfig 微信配置
type WeChatConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
}

// GetDSN 获取MySQL DSN
func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Database, d.Charset)
}

// GetAddr 获取Redis地址
func (r *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// GetServerAddr 获取服务器地址
func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// Load 加载配置
func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 环境变量支持
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// LoadDefault 加载默认配置
func LoadDefault() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "password",
			Database: "growth_tracker",
			Charset:  "utf8mb4",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
			DB:   0,
		},
		JWT: JWTConfig{
			Secret:     "growth-tracker-secret-key",
			ExpireTime: 86400 * 7, // 7天
		},
		AI: AIConfig{
			Provider:   "deepseek",
			Model:      "deepseek-chat",
			MaxTokens:  2000,
			Temperature: 0.7,
		},
		WeChat: WeChatConfig{
			AppID:     "",
			AppSecret: "",
		},
	}
}
