// config/override.go
package config

import (
	"flag"
	"fmt"
	"time"
)

// 从命令行参数覆盖配置
func (c *Config) OverrideByFlags() error {
	// 定义命令行参数（与配置结构体字段对应）
	serverAddr := flag.String("server-addr", "", "服务器地址（覆盖配置文件）")
	localAddr := flag.String("local-addr", "", "本地监听地址（覆盖配置文件）")
	logLevel := flag.String("log-level", "", "日志级别（覆盖配置文件）")
	reloadInterval := flag.String("reload-interval", "", "动态重载间隔（如 30s，覆盖配置文件）")

	flag.Parse()

	// 覆盖配置（仅当命令行参数非空时）
	if *serverAddr != "" {
		c.Client.ServerAddr = *serverAddr
	}
	if *localAddr != "" {
		c.Client.LocalAddr = *localAddr
	}
	if *logLevel != "" {
		c.LogLevel = *logLevel
	}
	if *reloadInterval != "" {
		dur, err := time.ParseDuration(*reloadInterval)
		if err != nil {
			return fmt.Errorf("无效的 reload-interval: %w", err)
		}
		c.ReloadInterval = dur
	}

	// 再次验证覆盖后的配置
	return c.Validate()
}
