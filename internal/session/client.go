package session

import (
	"clash_quic/config"
	"clash_quic/internal/security"
	"log"

	quic "github.com/quic-go/quic-go"
)

// 拨号 QUIC 服务器建立会话
func Dial(cfg *config.ClientConfig) (quic.Connection, error) {
	tlsConf := security.ClientTLSConfig(cfg)
	sess, err := quic.DialAddr(cfg.ServerAddr, tlsConf, nil)
	if err != nil {
		return nil, err
	}
	log.Printf("已连接到 QUIC 服务器: %s", cfg.ServerAddr)
	return sess, nil
}
