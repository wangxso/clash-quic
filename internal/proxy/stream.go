package proxy

import (
	"bufio"
	"log"
	"net"
	"sync"

	"clash_quic/internal/utils"

	quic "github.com/quic-go/quic-go"
)

// 处理服务器端 QUIC 流
func HandleStream(stream quic.Stream) {
	defer stream.Close()

	// 读取客户端发送的目标地址
	r := bufio.NewReader(stream)
	target, err := r.ReadString('\n')
	if err != nil {
		log.Printf("读取目标地址失败: %v", err)
		return
	}
	target = target[:len(target)-1] // 移除换行符
	log.Printf("处理目标地址: %s", target)

	// 连接目标服务器
	remoteConn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("连接目标服务器失败: %v", err)
		return
	}
	defer remoteConn.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	// 用 utils.Relay 替代手动 Copy（假设 Relay 函数需要两个 ReadWriter）
	go func() {
		defer wg.Done()
		utils.Relay(stream, remoteConn) // 本地连接 -> QUIC 流
	}()
	go func() {
		defer wg.Done()
		utils.Relay(remoteConn, stream) //  QUIC 流 -> 本地连接
	}()
	wg.Wait()
}
