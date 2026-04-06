package main

import (
	"io"
	"log"
	"net"
	"time"
)

const (
	ListenAddr       = "0.0.0.0:5190"
	OscarBackendAddr = "192.168.3.12:5190"
	OscarSig         = 0x2a
)

func handleConn(clientConn net.Conn) {
	defer clientConn.Close()

	clientConn.SetDeadline(time.Now().Add(2 * time.Second))
	buffer := make([]byte, 1)
	_, err := io.ReadFull(clientConn, buffer)
	if err != nil || buffer[0] != OscarSig {
		log.Printf("Non-OSCAR incoming connection=%v. Refuse it!", clientConn.RemoteAddr())
		return
	}

	clientConn.SetReadDeadline(time.Time{})

	backendConn, err := net.DialTimeout("tcp", OscarBackendAddr, 5*time.Second)
	if err != nil {
		log.Printf("Error connection to OSCAR backend: %v", err)
		return
	}
	defer backendConn.Close()

	backendConn.Write(buffer)

	done := make(chan bool, 2)

	go func() {
		io.Copy(backendConn, clientConn)
		done <- true
	}()

	go func() {
		io.Copy(clientConn, backendConn)
		done <- true
	}()

	<-done
}

func main() {
	listener, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("OSCAR Proxy started! %s -> %s", ListenAddr, OscarBackendAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConn(conn)
	}
}
