package proxy

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"clash_quic/internal/utils"

	quic "github.com/quic-go/quic-go"
)

// 处理本地 SOCKS5 连接
func HandleSocks5(localConn net.Conn, sess quic.Connection) {
	defer localConn.Close()

	// SOCKS5 握手
	if err := handshake(localConn); err != nil {
		log.Printf("SOCKS5 握手失败: %v", err)
		return
	}

	// 解析目标地址
	target, err := parseRequest(localConn)
	if err != nil {
		log.Printf("解析请求失败: %v", err)
		return
	}

	// 创建 QUIC 流并发送目标地址
	stream, err := sess.OpenStreamSync(context.Background())
	if err != nil {
		log.Printf("创建流失败: %v", err)
		localConn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0}) // 回复错误
		return
	}
	defer stream.Close()

	// 发送目标地址
	if _, err := stream.Write([]byte(target + "\n")); err != nil {
		log.Printf("发送目标地址失败: %v", err)
		return
	}

	// 回复客户端连接成功
	localConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// 转发数据
	utils.Relay(localConn, stream)
}

// SOCKS5 握手
func handshake(conn net.Conn) error {
	buf := make([]byte, 262)
	n, err := io.ReadAtLeast(conn, buf, 2)
	if err != nil {
		return err
	}

	if buf[0] != 0x05 { // 仅支持 SOCKS5
		return net.UnknownNetworkError("Just Support SOCKS5")
	}

	methodsLen := int(buf[1])
	need := 2 + methodsLen
	if n < need {
		_, err = io.ReadFull(conn, buf[n:need])
		if err != nil {
			return err
		}
	}

	// 回复：无认证
	_, err = conn.Write([]byte{0x05, 0x00})
	return err
}

// 解析 SOCKS5 请求获取目标地址
func parseRequest(conn net.Conn) (string, error) {
	buf := make([]byte, 262)
	_, err := io.ReadAtLeast(conn, buf, 4)
	if err != nil {
		return "", err
	}

	if buf[0] != 0x05 { // 版本错误
		return "", net.UnknownNetworkError("Version Error")
	}

	if buf[1] != 0x01 { // 仅支持 CONNECT 命令
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0}) // 不支持的命令
		return "", net.UnknownNetworkError("Command Error")
	}

	addrType := buf[3]
	idx := 4
	var host string

	switch addrType {
	case 0x01: // IPv4
		if _, err := io.ReadFull(conn, buf[idx:idx+4+2]); err != nil {
			return "", err
		}
		ip := net.IP(buf[idx : idx+4])
		port := binary.BigEndian.Uint16(buf[idx+4 : idx+6])
		host = net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))

	case 0x03: // 域名
		if _, err := io.ReadFull(conn, buf[idx:idx+1]); err != nil {
			return "", err
		}
		domainLen := int(buf[idx])
		idx++
		if _, err := io.ReadFull(conn, buf[idx:idx+domainLen+2]); err != nil {
			return "", err
		}
		domain := string(buf[idx : idx+domainLen])
		port := binary.BigEndian.Uint16(buf[idx+domainLen : idx+domainLen+2])
		host = net.JoinHostPort(domain, fmt.Sprintf("%d", port))

	case 0x04: // IPv6
		if _, err := io.ReadFull(conn, buf[idx:idx+16+2]); err != nil {
			return "", err
		}
		ip := net.IP(buf[idx : idx+16])
		port := binary.BigEndian.Uint16(buf[idx+16 : idx+18])
		host = net.JoinHostPort(ip.String(), fmt.Sprintf("%d", port))

	default:
		return "", net.UnknownNetworkError("Address Type Error")
	}

	return host, nil
}
