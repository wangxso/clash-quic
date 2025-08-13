package main

import (
	"clash_quic/config"
	"clash_quic/internal/proxy"
	"clash_quic/internal/session"
	"flag"
	"log"
	"net"
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
	sess, err := session.Dial(&cfg.Client)
	if err != nil {
		log.Fatalf("connect server failed: %v", err)
	}
	//启动socks5监听
	ln, err := net.Listen("tcp", cfg.Client.LocalAddr)
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}
	defer ln.Close()
	log.Printf("socks5 server listening on %s", cfg.Client.LocalAddr)
	// 处理本地连接
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed: %v", err)
			continue
		}
		go proxy.HandleSocks5(conn, sess)
	}

}
