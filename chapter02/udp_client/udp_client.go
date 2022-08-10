package client

import (
	"fmt"
	"net"
)

func UDPClient() {
	targetHost := "192.168.31.44"
	targetPort := ":9999"

	conn, err := net.Dial("udp", targetHost+targetPort)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	_, err = conn.Write([]byte("AAABBBCCC"))
	if err != nil {
		fmt.Println(err)
		return
	}

	var data [4096]byte
	_, err = conn.Read(data[:])
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(data[:]))
	fmt.Println(conn.RemoteAddr())
}
