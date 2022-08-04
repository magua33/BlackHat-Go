package main

import (
	"fmt"
	"unsafe"

	_ "golang.org/x/sys/windows"
)

const (
	IP_HDRINCL = 0x2
	SIO_RCVALL = windows.IOC_IN | windows.IOC_VENDOR | 1
	RCVALL_OFF = 0
	RCVALL_ON  = 1
)

var ADDR = [4]byte{192, 168, 31, 91}

func main() {
	var d windows.WSAData
	err := windows.WSAStartup(uint32(0x202), &d)
	if err != nil {
		fmt.Println("WSAStartup error:", err)
		return
	}

	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_RAW, windows.IPPROTO_IP)
	if err != nil {
		fmt.Println("Socket open error:", err)
		return
	}
	defer windows.Closesocket(fd)

	sa := windows.SockaddrInet4{
		Port: 0,
		Addr: ADDR,
	}

	err = windows.Bind(fd, &sa)
	if err != nil {
		fmt.Println("Socket bind error:", err)
		return
	}

	err = windows.SetsockoptInt(fd, windows.IPPROTO_IP, IP_HDRINCL, 1)
	if err != nil {
		fmt.Println("SetsockoptInt error:", err)
		return
	}

	unused := uint32(0)
	flag := uint32(RCVALL_ON)
	size := uint32(unsafe.Sizeof(flag))
	err = windows.WSAIoctl(fd, SIO_RCVALL, (*byte)(unsafe.Pointer(&flag)), size, nil, 0, &unused, nil, 0)
	if err != nil {
		fmt.Println("WSAIoctl error", err)
		return
	}

	data := make([]byte, 65535)
	n, from, err := windows.Recvfrom(fd, data, 0)
	if err != nil {
		fmt.Println("Windows Recvfrom:", err)
		return
	}

	addr := from.(*(windows.SockaddrInet4))
	host := fmt.Sprintf("%v.%v.%v.%v", addr.Addr[0], addr.Addr[1], addr.Addr[2], addr.Addr[3])
	fmt.Println("data:", string(data[:n]))
	fmt.Printf("host: %v port: %d \n", host, addr.Port)

	unused = uint32(0)
	flag = uint32(RCVALL_OFF)
	size = uint32(unsafe.Sizeof(flag))
	err = windows.WSAIoctl(fd, SIO_RCVALL, (*byte)(unsafe.Pointer(&flag)), size, nil, 0, &unused, nil, 0)
	if err != nil {
		fmt.Println("WSAIoctl error", err)
		return
	}
}
