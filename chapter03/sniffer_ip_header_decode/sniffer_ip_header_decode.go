package sniffer_ip_header_decode

import (
	"fmt"
	"syscall"

	"golang.org/x/net/ipv4"
)

// protocolMap 协议号-名称对照
var protocolMap = map[int]string{
	1:  "ICMP",
	6:  "IP",
	17: "UDP",
}

// protocol 返回协议的名字
func protocol(protocol int) string {
	if name, ok := protocolMap[protocol]; !ok {
		return fmt.Sprintf("No protocol for %d", protocol)
	} else {
		return name
	}
}

func sniff(host [4]byte) {
	// 创建原始套接字
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		fmt.Println("Init Socket Error:", err)
		return
	}
	defer syscall.Close(fd)

	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: host,
	}

	// 套接字绑定ip
	err = syscall.Bind(fd, &addr)
	if err != nil {
		fmt.Println("Socket Bind Error:", err)
		return
	}

	// 设置套接字捕获ip头
	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		fmt.Println("Set Socket Error:", err)
		return
	}

	for {
		buf := make([]byte, 65565)
		n, _, err := syscall.Recvfrom(fd, buf, 0)
		if err != nil {
			fmt.Println(err)
		}

		ipHeader, err := ipv4.ParseHeader(buf[:n])
		if err != nil {
			fmt.Println("ParseIPv4Header Error:", err)
		}

		if ipHeader != nil {
			fmt.Printf("Protocol: %s %v -> %v\n", protocol(ipHeader.Protocol), ipHeader.Src, ipHeader.Dst)
		}
	}
}
