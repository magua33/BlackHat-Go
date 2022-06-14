package server

import (
	"fmt"
	"net"
)

func UDPServer() {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 9999,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	defer conn.Close()

	for {
		var data [1024]byte
		n, addr, err := conn.ReadFromUDP(data[:])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("data:", data)
		fmt.Println("len:", n)
		fmt.Println("addr", addr)

		_, err = conn.WriteToUDP(data[:n], addr)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
