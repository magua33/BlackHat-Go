package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args[1:]) != 5 {
		fmt.Println("Usage ./proxy.go [LocalHost] [LocalPort]")
		fmt.Println("[RemoteHost][RemotePort] [ReceiveFirst]")
		fmt.Println("Example ./proxy.go 127.0.0.1 9000 12.12.132.1 9000 True")
		os.Exit(0)
	}

	localHost := os.Args[1]
	localPort := os.Args[2]

	remoteHost := os.Args[3]
	remotePort := os.Args[4]

	receiveFirst := strings.Contains(os.Args[5], "True")

	fmt.Println(localHost, localPort, remoteHost, remotePort, receiveFirst)

	serverLoop(localHost, localPort, remoteHost, remotePort, receiveFirst)
}

func hexDump(data string) string {
	src := []rune(data)
	length := 16
	show := true

	var result string
	for i := 0; i < len(src); i += length {
		word := []rune{}
		if i+length > len(src) {
			word = src[i:]
		} else {
			word = src[i : i+length]
		}

		printable := strings.Map(func(r rune) rune {
			if !strconv.IsPrint(r) {
				return '.'
			}
			return r
		}, string(word))

		hexA := strings.Trim(fmt.Sprintf("%02X", word), "[]")
		hexWidth := length * 4
		line := fmt.Sprintf("%04x \t %s %*s  %s\n", i, hexA, hexWidth-len(hexA), "-", printable)
		result += line
	}

	if show {
		fmt.Println(result)
	}
	return result
}

func receiveFrom(conn net.Conn) []byte {
	buffer := []byte{}

	conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	for {
		data := [4096]byte{}
		n, err := conn.Read(data[:])
		if err != nil && err == io.EOF {
			break
		}
		if n == 0 {
			break
		}
		buffer = append(buffer, data[:n]...)
	}

	return buffer
}

func requestHandler(buffer []byte) []byte {
	// perform packet modifications
	return buffer
}

func responseHandler(buffer []byte) []byte {
	// perform packet modifications
	return buffer
}

func serverLoop(localHost, localPort, remoteHost, remotePort string, receiveFirst bool) {
	server, err := net.Listen("tcp", localHost+":"+localPort)
	if err != nil {
		fmt.Println("problem on bind:", err.Error())
		fmt.Printf("[!!] Failed to listen on %s:%s\n", localHost, localPort)
		fmt.Println("[!!] Check for the listening sockets or correct permissions.")
		os.Exit(0)
	}

	defer server.Close()
	fmt.Printf("[*] Listening on %s:%s\n", localHost, localPort)

	for {
		clientSocket, err := server.Accept()
		if err != nil {
			fmt.Println("Server Accept Error:", err.Error())
			continue
		}

		go func(clientSocket net.Conn) {
			line := "> Received incoming connection from " + clientSocket.RemoteAddr().String()
			fmt.Println(line)

			proxyHandler(clientSocket, remoteHost, remotePort, receiveFirst)
		}(clientSocket)

	}
}

func proxyHandler(clientSocket net.Conn, remoteHost, remotePort string, receiveFirst bool) {
	remoteSocket, err := net.Dial("tcp", remoteHost+":"+remotePort)
	if err != nil {
		fmt.Println("Net Dial Error:", err.Error())
		return
	}

	if receiveFirst {
		remoteBuffer := receiveFrom(remoteSocket)
		hexDump(string(remoteBuffer))

		remoteBuffer = responseHandler(remoteBuffer)
		if len(remoteBuffer) > 0 {
			fmt.Printf("[<==] Sending %d bytes to localhost.\n", len(remoteBuffer))
			clientSocket.Write(remoteBuffer)
		}
	}

	for {
		localBuffer := receiveFrom(clientSocket)
		if len(localBuffer) > 0 {
			line := fmt.Sprintf("[==>] Received %d bytes from localhost.", len(localBuffer))
			fmt.Println(line)
			hexDump(string(localBuffer))

			localBuffer = requestHandler(localBuffer)
			remoteSocket.Write(localBuffer)
			fmt.Println("[==>] Send to remote.")
		}

		remoteBuffer := receiveFrom(remoteSocket)
		if len(remoteBuffer) > 0 {
			fmt.Printf("[<==] Received %d bytes from remote.\n", len(remoteBuffer))
			hexDump(string(remoteBuffer))

			remoteBuffer = responseHandler(remoteBuffer)
			clientSocket.Write(remoteBuffer)
			fmt.Println("[<==] Send to localhost.")
		}

		if len(localBuffer) == 0 || len(remoteBuffer) == 0 {
			clientSocket.Close()
			remoteSocket.Close()
			fmt.Println("[*] No more data. Closing connections.")
			break
		}
	}
}
