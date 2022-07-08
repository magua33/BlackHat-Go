package main

import (
	"bufio"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() {
	var ip, port, user, passwd, cmd string

	fmt.Print("Username: ")
	fmt.Scan(&user)

	fmt.Print("Password: ")
	bytepw, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Println("ReadPassword Error:", err)
		return
	}
	passwd = string(bytepw)
	fmt.Println()

	fmt.Print("Enter server IP: ")
	fmt.Scan(&ip)

	fmt.Print("Enter port or <CR>: ")
	fmt.Scanln(&port)
	if port == "" {
		port = "22"
	}

	fmt.Print("Enter command or <CR>: ")
	reader := bufio.NewReader(os.Stdin)
	cmd, _ = reader.ReadString('\n')

	fmt.Println("ip:", ip, "port:", port, "user:", user, "passwd:", "******", "cmd:", cmd)

	sshCommand(ip, port, user, passwd, cmd)
}

func sshCommand(ip, port, user, passwd, cmd string) {
	conf := ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
	}

	addr := ip + ":" + port
	client, err := ssh.Dial("tcp", addr, &conf)
	if err != nil {
		fmt.Println("SSH Dial Error:", err)
		return
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Println("Client NewSession Error:", err)
		return
	}
	defer session.Close()

	comb, err := session.CombinedOutput(cmd)
	if err != nil {
		fmt.Println("Exec Command Error:", err)
	}

	fmt.Println("--- Output ---")
	fmt.Print(string(comb))
}
