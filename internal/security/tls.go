// TLS配置

package security

import (
	"clash_quic/config"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
)

func ServerTLSConfig(certFile, keyFile string) *tls.Config {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("加载证书失败: %v", err)
	}
	// 检测证书是否配置正确
	if len(cert.Certificate) == 0 {
		log.Fatal("证书文件未配置正确")
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-socks5-clash"},
	}
}

func ClientTLSConfig(cfg *config.ClientConfig) *tls.Config {
	tlsConf := &tls.Config{
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		NextProtos:         []string{"quic-socks5-clash"},
	}
	// 支持加载自定义CA证书（若配置了则优先使用）
	if cfg.CACertPath != "" {
		caCert, err := os.ReadFile(cfg.CACertPath)
		if err != nil {
			log.Printf("read ca cert err: %v", err)
			return tlsConf
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConf.RootCAs = caCertPool
	}
	return tlsConf
}
