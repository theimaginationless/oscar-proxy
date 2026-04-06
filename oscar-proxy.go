package main

import (
	"bytes"
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

var oscarHello = []byte{0x2a, 0x01, 0x00, 0x01, 0x00, 0x04, 0x00, 0x00, 0x00, 0x01}

func handleConn(clientConn net.Conn) {
	defer clientConn.Close()

	clientConn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	_, err := clientConn.Write(oscarHello)
	if err != nil {
		return
	}

	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))

	buffer := make([]byte, 1)
	_, err = io.ReadFull(clientConn, buffer)
	if err != nil || buffer[0] != OscarSig {
		log.Printf("Non-OSCAR incoming connection=%v. Data=%v", clientConn.RemoteAddr(), buffer)
		return
	}

	clientConn.SetDeadline(time.Time{})

	backendConn, err := net.DialTimeout("tcp", OscarBackendAddr, 5*time.Second)
	if err != nil {
		log.Printf("Error connection to OSCAR backend: %v", err)
		return
	}
	defer backendConn.Close()

	log.Printf("Initial connection... ip=%v", clientConn.RemoteAddr())

	clientStream := io.MultiReader(bytes.NewReader(buffer), clientConn)

	done := make(chan bool, 2)

	go func() {
		_, err := io.Copy(backendConn, clientStream)
		if err != nil {
			log.Printf("Copy to backend error: %v", err)
		}
		done <- true
	}()

	go func() {
		_, err := io.Copy(clientConn, backendConn)
		if err != nil {
			log.Printf("Copy to client error: %v", err)
		}
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
