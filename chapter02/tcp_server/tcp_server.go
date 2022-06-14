package main

import (
	"fmt"
	"net"

	"golang.org/x/net/netutil"
)

func main() {
	TCPServer()
}

var IP = "0.0.0.0"
var Port = ":9999"

func TCPServer() {
	listener, err := net.Listen("tcp", IP+Port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	fmt.Printf("[*] Listening on %s%s\n", IP, Port)

	listener = netutil.LimitListener(listener, 5)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go func(conn net.Conn) {
			defer conn.Close()

			fmt.Printf("[*] Accepted connection form %s\n", conn.RemoteAddr().String())

			var request [1024]byte
			_, err := conn.Read(request[:])
			if err != nil {
				fmt.Println(err)
				return
			}

			conn.Write([]byte("hello"))
			fmt.Printf("[*] Received: %s\n", string(request[:]))
			conn.Write([]byte("ACK"))
		}(conn)
	}
}
