package sniffer

import (
	"fmt"

	"golang.org/x/net/icmp"
)

// sniffer 一个简易的嗅探器
// host 监听的主机ip
func sniffer(host string) {
	// 侦听传入的icmp数据包
	conn, err := icmp.ListenPacket("ip4:icmp", host)
	if err != nil {
		fmt.Println("icmp listen error:", err)
		return
	}

	// 读取一个数据包
	buf := make([]byte, 1024)
	n, addr, err := conn.ReadFrom(buf)
	if err != nil {
		fmt.Println("recv error:", err)
		return
	}
	fmt.Println(string(buf[:n]))
	fmt.Println(addr)
}

/*
func main() {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		fmt.Println("Init Socket Error:", err)
		return
	}
	defer syscall.Close(fd)

	addr := syscall.SockaddrInet4{
		Port: 0,
		Addr: [4]byte{0, 0, 0, 0},
	}
	err = syscall.Bind(fd, &addr)
	if err != nil {
		fmt.Println("Socket Bind Error:", err)
		return
	}

	err = syscall.SetsockoptInt(fd, syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	if err != nil {
		fmt.Println("Set Socket Error:", err)
		return
	}

	buf := make([]byte, 65565)
	n, from, err := syscall.Recvfrom(fd, buf, 0)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(string(buf[:n]))
	fmt.Println(toStr(from))
}

func toStr(addr syscall.Sockaddr) string {
	b := addr.(*syscall.SockaddrInet4).Addr
	return fmt.Sprintf("%v.%v.%v.%v", b[0], b[1], b[2], b[3])
}
*/
