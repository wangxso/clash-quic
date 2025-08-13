// config/parser.go
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// 从文件加载配置（支持相对路径和绝对路径）
func LoadFromFile(path string) (*Config, error) {
	// 读取文件内容
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析 YAML 到结构体（基于默认配置合并，避免配置文件缺失字段）
	cfg := Default()
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置合法性（如必填字段、格式校验）
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("配置不合法: %w", err)
	}

	return cfg, nil
}

// 配置验证（确保关键字段有效）
func (c *Config) Validate() error {
	if c.Mode != "client" && c.Mode != "server" {
		return fmt.Errorf("mode 必须为 'client' 或 'server'")
	}

	if c.Mode == "client" {
		if c.Client.ServerAddr == "" {
			return fmt.Errorf("客户端配置中 server-addr 不能为空")
		}
	} else {
		if c.Server.CertPath == "" || c.Server.KeyPath == "" {
			return fmt.Errorf("服务器配置中 cert-path 和 key-path 不能为空")
		}
		// 检查证书文件是否存在
		if _, err := os.Stat(c.Server.CertPath); err != nil {
			return fmt.Errorf("证书文件不存在: %s", c.Server.CertPath)
		}
	}

	return nil
}
