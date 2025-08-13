package session

import (
	"clash_quic/internal/security"
	"context"
	"log"

	quic "github.com/quic-go/quic-go"
)

// 启动 QUIC 服务器监听
func Listen(addr, certFile, keyFile string) (quic.Listener, error) {
	tlsConf := security.ServerTLSConfig(certFile, keyFile)
	listener, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("QUIC 服务器监听: %s", addr)
	return listener, nil
}

// 处理 QUIC 会话（接受流）
func HandleSession(sess quic.Connection, streamHandler func(quic.Stream)) {
	defer sess.CloseWithError(0, "会话结束")
	log.Printf("接受来自 %v 的连接", sess.RemoteAddr())

	for {
		stream, err := sess.AcceptStream(context.Background())
		if err != nil {
			log.Printf("接受流失败: %v", err)
			return
		}
		go streamHandler(stream)
	}
}
