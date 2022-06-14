package main

import (
	"fmt"
	"net"
)

func main() {
	TCPClient("www.baidu.com:80", "")
}

func TCPClient(addr, data string) {
	conn, _ := net.Dial("tcp", addr)

	defer conn.Close()

	conn.Write([]byte(data))

	var buf [4096]byte
	conn.Read(buf[:])

	fmt.Println(string(buf[:]))
}
