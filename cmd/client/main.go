package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	quic "github.com/quic-go/quic-go"
)

var (
	serverAddr = "127.0.0.1:4242" // remote quic server
	localAddr  = "127.0.0.1:1080" // local socks5 listener
)

func clientTLSConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true, // demo only
		NextProtos:         []string{"quic-socks5-demo"},
	}
}

func handleConn(localConn net.Conn, sess quic.Connection) {
	defer localConn.Close()

	// Minimal SOCKS5 handling: only support NO-AUTH and CONNECT
	buf := make([]byte, 262)
	// handshake
	n, err := io.ReadAtLeast(localConn, buf, 2)
	if err != nil {
		log.Printf("handshake read err: %v", err)
		return
	}
	if buf[0] != 0x05 {
		log.Printf("not socks5")
		return
	}
	methods := int(buf[1])
	need := 2 + methods
	if n < need {
		_, err = io.ReadFull(localConn, buf[n:need])
		if err != nil {
			log.Printf("read methods err: %v", err)
			return
		}
	}
	// reply: no auth
	localConn.Write([]byte{0x05, 0x00})

	// request
	n, err = io.ReadAtLeast(localConn, buf, 4)
	if err != nil {
		log.Printf("request read err: %v", err)
		return
	}
	if buf[0] != 0x05 {
		log.Printf("invalid req ver")
		return
	}
	cmd := buf[1]
	if cmd != 0x01 {
		log.Printf("only CONNECT supported")
		localConn.Write([]byte{0x05, 0x07, 0x00, 0x01, 0, 0, 0, 0, 0, 0}) // host unreachable
		return
	}
	addrType := buf[3]
	idx := 4
	var host string
	switch addrType {
	case 0x01: // IPv4
		_, err = io.ReadFull(localConn, buf[idx:idx+4+2])
		if err != nil {
			log.Printf("read ipv4 err: %v", err)
			return
		}
		host = net.IP(buf[idx : idx+4]).String()
		port := binary.BigEndian.Uint16(buf[idx+4 : idx+6])
		host = net.JoinHostPort(host, fmt.Sprintf("%d", port))
	case 0x03: // Domain
		// read domain length
		_, err = io.ReadFull(localConn, buf[idx:idx+1])
		if err != nil {
			log.Printf("read domain len err: %v", err)
			return
		}
		dlen := int(buf[idx])
		_, err = io.ReadFull(localConn, buf[idx+1:idx+1+dlen+2])
		if err != nil {
			log.Printf("read domain err: %v", err)
			return
		}
		domain := string(buf[idx+1 : idx+1+dlen])
		port := binary.BigEndian.Uint16(buf[idx+1+dlen : idx+1+dlen+2])
		host = net.JoinHostPort(domain, fmt.Sprintf("%d", port))
	case 0x04: // IPv6
		_, err = io.ReadFull(localConn, buf[idx:idx+16+2])
		if err != nil {
			log.Printf("read ipv6 err: %v", err)
			return
		}
		host = net.IP(buf[idx : idx+16]).String()
		port := binary.BigEndian.Uint16(buf[idx+16 : idx+18])
		host = net.JoinHostPort(host, fmt.Sprintf("%d", port))
	default:
		log.Printf("unknown addr type %d", addrType)
		return
	}

	// open a quic stream to server and send target addr (text line)
	stream, err := sess.OpenStreamSync(context.Background())
	if err != nil {
		log.Printf("open stream err: %v", err)
		localConn.Write([]byte{0x05, 0x01, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		return
	}
	// send target + '\n'
	stream.Write([]byte(host + "\n"))

	// success reply to client: VER, REP=0x00, RSV, ATYP(IPv4 loopback), BND.ADDR, BND.PORT
	// we just reply with a dummy IPv4 0.0.0.0:0
	localConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0, 0, 0, 0, 0, 0})

	// pipe
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		io.Copy(stream, localConn)
		stream.Close()
	}()
	go func() {
		defer wg.Done()
		io.Copy(localConn, stream)
		localConn.Close()
	}()
	wg.Wait()
}

func main() {
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}
	// establish a QUIC session and reuse it
	tlsConf := clientTLSConfig()
	sess, err := quic.DialAddr(serverAddr, tlsConf, nil)
	if err != nil {
		log.Fatalf("dial quic server failed: %v", err)
	}
	log.Printf("connected to quic server %s", serverAddr)

	ln, err := net.Listen("tcp", localAddr)
	if err != nil {
		log.Fatalf("listen failed: %v", err)
	}
	log.Printf("listening SOCKS5 on %s", localAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept local err: %v", err)
			continue
		}
		go handleConn(conn, sess)
	}
}
