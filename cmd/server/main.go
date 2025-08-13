package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"os"

	quic "github.com/quic-go/quic-go"
)

func tlsConfig() *tls.Config {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		log.Fatalf("failed to load cert: %v", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-socks5-demo"},
	}
}

func handleStream(s quic.Stream) {
	defer s.Close()
	// First line from client is target address, e.g. "example.com:80\n"
	r := bufio.NewReader(s)
	target, err := r.ReadString('\n')
	if err != nil {
		log.Printf("failed read target: %v", err)
		return
	}
	target = target[:len(target)-1] // strip '\n'
	log.Printf("stream for target %s", target)

	// Dial target TCP
	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Printf("dial target %s failed: %v", target, err)
		// Optionally send error proto back
		return
	}
	defer conn.Close()

	// Now we have two connections: s <-> conn
	// Note: r already buffered some bytes from the stream; flush by copying remaining buffer
	// Create a pipe between the two
	go func() {
		_, err := io.Copy(conn, r)
		if err != nil {
			// log.Printf("copy to target err: %v", err)
		}
		conn.Close()
	}()

	_, err = io.Copy(s, conn)
	if err != nil {
		// log.Printf("copy to stream err: %v", err)
	}
}

func main() {
	if _, err := os.Stat("cert.pem"); os.IsNotExist(err) {
		log.Fatalf("cert.pem not found. run gen_cert.sh first")
	}

	addr := ":4242"
	log.Printf("listening QUIC on %s", addr)
	listener, err := quic.ListenAddr(addr, tlsConfig(), nil)
	if err != nil {
		log.Fatalf("quic listen failed: %v", err)
	}
	for {
		sess, err := listener.Accept(context.Background())
		if err != nil {
			log.Printf("accept conn err: %v", err)
			continue
		}
		log.Printf("accepted connection from %v", sess.RemoteAddr())
		// accept streams from session
		go func(sess quic.Connection) {
			for {
				stream, err := sess.AcceptStream(context.Background())
				if err != nil {
					log.Printf("accept stream err: %v", err)
					return
				}
				go handleStream(stream)
			}
		}(sess)
	}
}
