package main

import (
	"fmt"
	"io"
	"net"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

type options struct {
	Password string
	User     string
	Port     string
}

func main() {
	options := options{
		Password: "kali",
		User:     "kali",
		Port:     "8081",
	}
	serverAddr, remoteAddr := "192.168.31.183:22", "192.168.31.44:3000"

	if options.Password == "" {
		fmt.Print("Enter SSH password: ")
		bytePw, _ := term.ReadPassword(int(syscall.Stdin))
		options.Password = string(bytePw)
		fmt.Println()
	}

	conf := ssh.ClientConfig{
		User:            options.User,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(options.Password)},
	}

	local, err := net.Listen("tcp", "127.0.0.1:"+options.Port)
	if err != nil {
		fmt.Println("C-c: Port forwarding stopped. Error:", err.Error())
		return
	}
	defer local.Close()

	fmt.Printf("Listening on port: %s\n", options.Port)

	for {
		client, err := local.Accept()
		if err != nil {
			fmt.Println("Local Accept Error:", err.Error())
			continue
		}

		go reverseForwardTunnel(serverAddr, remoteAddr, options.Port, conf, client)
	}
}

func reverseForwardTunnel(serverAddr, remoteAddr, localPort string, conf ssh.ClientConfig, client net.Conn) {
	sshConn, err := ssh.Dial("tcp", serverAddr, &conf)
	if err != nil {
		fmt.Printf("*** Failed to connect to %s, Error: %s\n", serverAddr, err.Error())
		return
	}
	fmt.Printf("Connecting to host %s ...\n", serverAddr)

	remote, err := sshConn.Dial("tcp", remoteAddr)
	if err != nil {
		fmt.Printf("*** Failed to connect to %s, Error: %s \n", remoteAddr, err.Error())
		return
	}

	fmt.Printf("Now forwarding remote port %s to %s ...\n", localPort, remoteAddr)
	go forward(client, remote, remoteAddr)
}

func forward(client, remote net.Conn, remoteAddr string) {
	chDone := make(chan bool)

	fmt.Printf("Connected! Tunnel open (%s) -> (%s) -> (%s)\n", client.RemoteAddr(), client.LocalAddr(), remoteAddr)
	go func() {
		_, err := io.Copy(client, remote)
		if err != nil {
			fmt.Println("Error while copy remote -> local:", err.Error())
		}
		chDone <- true
	}()

	go func() {
		_, err := io.Copy(remote, client)
		if err != nil {
			fmt.Println("Error while copy local -> remote:", err.Error())
		}
		chDone <- true
	}()

	<-chDone
}
