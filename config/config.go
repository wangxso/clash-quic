// config/config.go
package config

import (
	"time"
)

// 全局配置结构体（区分客户端/服务器模式）
type Config struct {
	Mode           string        `yaml:"mode"` // "client" 或 "server"
	Client         ClientConfig  `yaml:"client"`
	Server         ServerConfig  `yaml:"server"`
	LogLevel       string        `yaml:"log-level"`       // "debug", "info", "error"
	LogFile        string        `yaml:"log-file"`        // 日志文件路径，空则输出到控制台
	ReloadInterval time.Duration `yaml:"reload-interval"` // 动态重载间隔（如 30s）
}

// 客户端配置
type ClientConfig struct {
	ServerAddr     string        `yaml:"server-addr"`     // 服务器地址（如 "example.com:443"）
	LocalAddr      string        `yaml:"local-addr"`      // 本地监听地址（如 "127.0.0.1:1080"）
	TLSEnable      bool          `yaml:"tls-enable"`      // 是否启用 TLS 验证
	CACertPath     string        `yaml:"ca-cert-path"`    // CA 证书路径
	AuthToken      string        `yaml:"auth-token"`      // 认证令牌
	DialTimeout    time.Duration `yaml:"dial-timeout"`    // 连接超时（如 5s）
	ReconnectTimes int           `yaml:"reconnect-times"` // 重连次数
}

// 服务器配置
type ServerConfig struct {
	ListenAddr  string        `yaml:"listen-addr"`  // 监听地址（如 ":443"）
	CertPath    string        `yaml:"cert-path"`    // 服务器证书路径
	KeyPath     string        `yaml:"key-path"`     // 私钥路径
	MaxStreams  int           `yaml:"max-streams"`  // 最大并发流数
	AllowedIPs  []string      `yaml:"allowed-ips"`  // 允许连接的客户端 IP（空表示不限制）
	AuthTokens  []string      `yaml:"auth-tokens"`  // 允许的认证令牌列表（空表示不认证）
	ReadTimeout time.Duration `yaml:"read-timeout"` // 读超时
}

// 默认配置（确保每个字段都有合理默认值）
func Default() *Config {
	return &Config{
		Mode:           "client",
		LogLevel:       "info",
		ReloadInterval: 30 * time.Second,
		Client: ClientConfig{
			ServerAddr:     "127.0.0.1:443",
			LocalAddr:      "127.0.0.1:1080",
			TLSEnable:      true,
			DialTimeout:    5 * time.Second,
			ReconnectTimes: 3,
		},
		Server: ServerConfig{
			ListenAddr:  ":443",
			CertPath:    "cert.pem",
			KeyPath:     "key.pem",
			MaxStreams:  100,
			ReadTimeout: 30 * time.Second,
		},
	}
}
