package main

import (
	"clash_quic/config"
	"clash_quic/internal/proxy"
	"clash_quic/internal/session"
	"context"
	"flag"
	"log"
)

func main() {
	// 初始化配置管理器
	configPath := flag.String("config", "config/config.yaml", "配置文件路径")
	flag.Parse()

	mgr, err := config.NewManager(*configPath)
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}
	defer mgr.Stop()
	// 从管理器获取配置
	cfg := mgr.Get()
	log.Printf("使用配置: %+v", cfg.Client)
	// 启动 QUIC 监听
	ln, err := session.Listen(cfg.Server.ListenAddr, cfg.Server.CertPath, cfg.Server.KeyPath)
	if err != nil {
		log.Fatalf("QUIC 监听失败: %v", err)
	}
	defer ln.Close()
	// 接受会话并处理
	for {
		sess, err := ln.Accept(context.Background())
		if err != nil {
			log.Printf("接受会话失败: %v", err)
			continue
		}
		go session.HandleSession(sess, proxy.HandleStream)
	}
}
