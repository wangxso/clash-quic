package proxy

import (
	"io"
	"log"

	quic "github.com/quic-go/quic-go"
	"github.com/shadowsocks/go-shadowsocks2"
)

// 处理Shadowsocks客户端流（加密发送数据）
func HandleShadowsocksClientStream(stream quic.Stream, method, password string) error {
	// 创建Shadowsocks密码器
	ciph, err := shadowsocks.NewCipher(method, password)
	if err != nil {
		return err
	}

	// 创建加密写入器
	writer := ciph.NewWriter(stream)

	// 读取目标地址并加密发送
	target, err := io.ReadAll(stream)
	if err != nil {
		return err
	}
	_, err = writer.Write(target)
	return err
}

// 处理Shadowsocks服务器流（解密接收数据）
func HandleShadowsocksServerStream(stream quic.Stream, method, password string) error {
	// 创建Shadowsocks密码器
	ciph, err := shadowsocks.NewCipher(method, password)
	if err != nil {
		return err
	}

	// 创建解密读取器
	reader := ciph.NewReader(stream)

	// 解密读取目标地址
	target, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	log.Printf("解密后目标地址: %s", target)

	// 连接目标服务器并转发数据（与stream.go逻辑类似）
	// TODO: 实现与stream.go中类似的目标连接和数据转发逻辑
	return nil
}
