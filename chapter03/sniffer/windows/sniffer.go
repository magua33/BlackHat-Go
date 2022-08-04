package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const SIO_RCVALL = windows.IOC_IN | windows.IOC_VENDOR | 1

func main() {
	var wsaData windows.WSAData
	err := windows.WSAStartup(2<<16+2, &wsaData)
	if err != nil {
		fmt.Println("Windows WSAStartup:", err)
		return
	}

	fd, err := windows.Socket(windows.AF_INET, windows.SOCK_RAW, windows.IPPROTO_IP)
	if err != nil {
		fmt.Println("Windows Socket:", err)
		return
	}

	la := new(windows.SockaddrInet4)
	la.Port = int(0)
	err = windows.Bind(fd, la)
	if err != nil {
		fmt.Println("Windows Bind:", err)
		return
	}

	// var aOptVal bool = true
	// err = windows.Setsockopt(fd, windows.IPPROTO_IP, windows.IP_HDRINCL, (*byte)(unsafe.Pointer(&aOptVal)), int32(unsafe.Sizeof(aOptVal)))
	err = windows.SetsockoptInt(fd, windows.IPPROTO_IP, windows.IP_HDRINCL, 1)
	if err != nil {
		fmt.Println("Windows Setsockopt:", err)
		return
	}

	inbuf := uint32(1)
	sizebuf := uint32(unsafe.Sizeof(inbuf))
	ret := uint32(0)
	err = windows.WSAIoctl(fd, SIO_RCVALL, (*byte)(unsafe.Pointer(&inbuf)), sizebuf, nil, 0, &ret, nil, 0)
	if err != nil {
		fmt.Println("Windows WSAIoctl:", err)
	}

	data := make([]byte, 65535)
	n, addr, err := windows.Recvfrom(fd, data, 0)
	if err != nil {
		fmt.Println("Windows Recvfrom:", err)
	}
	fmt.Println("data:", string(data[:n]))
	fmt.Println("addr:", addr)

	// 关闭
	err = windows.Closesocket(fd)
	if err != nil {
		fmt.Println("Windows Closesocket:", err)
		return
	}
	err = windows.WSACleanup()
	if err != nil {
		fmt.Println("Windows WSACleanup:", err)
		return
	}
}
