package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

func main() {
	conf := &ssh.ServerConfig{
		// Remove to disable password auth.
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			// Should use constant-time compare (or better, salt+hash) in
			// a production setting.
			if c.User() == "tim" && string(pass) == "secret" {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},

		ServerVersion: "SSH-2.0-OWN-SERVER",
	}

	pemBytes, err := ioutil.ReadFile("test_rsa.key")
	if err != nil {
		fmt.Println("ioutil ReadFile Error:", err.Error())
		return
	}

	privateKeys, err := ssh.ParsePrivateKey(pemBytes)
	if err != nil {
		fmt.Println("ssh ParsePrivateKey Error:", err.Error())
		return
	}

	conf.AddHostKey(privateKeys)

	server := "0.0.0.0"
	port := "2222"
	addr := server + ":" + port

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("[-] Listen failed:", err)
		return
	}
	defer listener.Close()

	fmt.Println("[+] Listening for connection ...")

	tcpConn, err := listener.Accept()
	if err != nil {
		fmt.Println("listener Accept Error:", err)
		return
	}

	fmt.Println("[+] Got a connection!", tcpConn.RemoteAddr().String())

	_, chans, reqs, err := ssh.NewServerConn(tcpConn, conf)
	if err != nil {
		fmt.Println("ssh NewServerConn Error:", err)
		return
	}

	go ssh.DiscardRequests(reqs)

	exit := make(chan bool)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, _, err := newChannel.Accept()
		if err != nil {
			fmt.Println("newChannel Accept Error:", err)
		}

		fmt.Println("[+] Authenticated!")

		var recv [1024]byte
		n, err := channel.Read(recv[:])
		if err != nil {
			fmt.Println("Channel Read Error:", err)
		}

		fmt.Println(">", string(recv[:n]))
		channel.Write([]byte("Welcome to bh_ssh"))

		for {
			fmt.Print("Enter command: ")
			read := bufio.NewReader(os.Stdin)
			command, err := read.ReadString('\n')
			if err != nil {
				fmt.Println("Enter Command Error:", err)
				continue
			}
			command = strings.TrimSpace(command)
			fmt.Println("<", command)

			if command != "exit" {
				_, err := channel.Write([]byte(command))
				if err != nil {
					fmt.Println("Channel Write Error:", err)
					continue
				}

				var recv [8192]byte
				n, err := channel.Read(recv[:])
				if err != nil {
					fmt.Println("Channel Read Error:", err)
					continue
				}
				fmt.Println(">", string(recv[:n]))
			} else {
				channel.Write([]byte("exit"))
				fmt.Println("exiting")
				exit <- true
				break
			}

		}
	}

	<-exit
}
