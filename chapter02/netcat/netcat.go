package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	// 使用方法 帮助信息
	flag.Usage = func() {
		help := "usage: netcat.go [-h] [-c] [-e EXECUTE] [-l] [-p PORT] [-t TARGET] [-u UPLOAD]\n\n" +
			"BHGo Net Tool\n\n" +

			"optional arguments:" +
			"  -h, --help            show this help message and exit\n" +
			"  -c, --command         command shell\n" +
			"  -e EXECUTE, --execute EXECUTE\n" +
			"                        execute specified command\n" +
			"  -l, --listen          listen\n" +
			"  -p PORT, --port PORT  specified port\n" +
			"  -t TARGET, --target TARGET\n" +
			"                        specified IP\n" +
			"  -u UPLOAD, --upload UPLOAD\n" +
			"                        upload file\n\n" +

			"Example:\n" +
			"netcat.go -t 192.168.1.108 -p 5555 -l -c # command shell\n" +
			"netcat.go -t 192.168.1.108 -p 5555 -l -u=mytest.txt # upload to file\n" +
			"netcat.go -t 192.168.1.108 -p 5555 -l -e=\"cat /etc/passwd\" # execute command\n" +
			"echo 'ABCDEFGHI' | ./netcat.go -t 192.168.11.12 -p 135 # echo text to server port 135\n" +
			"netcat.go -t 192.168.1.108 -p 5555 connect to server"
		fmt.Println(help)
	}

	// 解析参数
	parser := parser{}
	flag.BoolVar(&parser.command, "c", false, "command shell")
	flag.StringVar(&parser.execute, "e", "", "execute specified command")
	flag.BoolVar(&parser.listen, "l", false, "listen")
	flag.StringVar(&parser.port, "p", "", "specified port")
	flag.StringVar(&parser.target, "t", "", "specified IP")
	flag.StringVar(&parser.upload, "u", "", "upload file")
	flag.Parse()
	fmt.Printf("%+v\n", parser)

	var buffer []byte
	if parser.listen {
		buffer = nil
	} else {
		// 读取标准输入数据
		reader := bufio.NewReader(os.Stdin)
		for {
			b, err := reader.ReadByte()
			if err != nil {
				break
			}
			buffer = append(buffer, b)
		}
	}

	// 初始化netcat
	nc := initNetcat(parser, buffer)
	nc.run()
}

type netcat struct {
	parser parser
	buffer []byte
	conn   net.Conn
}

type parser struct {
	command bool
	execute string
	listen  bool
	port    string
	target  string
	upload  string
}

// execute 执行传入命令
func execute(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return ""
	}

	cmdAndArgs := strings.Split(cmd, " ")
	command := exec.Command(cmdAndArgs[0], cmdAndArgs[1:]...)

	// 获取标准输出
	closer, err := command.StdoutPipe()
	if err != nil {
		fmt.Println(err)
	}

	command.Start()

	// 从标准输出中获取命令执行结果
	rsp, err := io.ReadAll(closer)
	if err != nil {
		fmt.Println(err)
	}

	command.Wait()

	return string(rsp)
}

func initNetcat(parser parser, buffer []byte) netcat {
	return netcat{
		parser: parser,
		buffer: buffer,
	}
}

func (nc *netcat) run() {
	if nc.parser.listen {
		nc.listen()
	} else {
		nc.send()
	}
}

func (nc *netcat) send() {
	addr := nc.parser.target + ":" + nc.parser.port
	var err error
	nc.conn, err = net.Dial("tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer nc.conn.Close()

	// 如果缓冲区有数据，则发送
	if nc.buffer != nil {
		_, err := nc.conn.Write(nc.buffer)
		if err != nil {
			fmt.Println(err)
		}
	}

	// 接收target返回的数据
	for {
		recvLen := 1
		response := []byte{}

		// 读取数据
		for recvLen > 0 {
			var buf [4096]byte
			n, err := nc.conn.Read(buf[:])
			if err != nil {
				fmt.Println(err)
				break
			}

			recvLen = n
			response = append(response, buf[:]...)
			if recvLen < 4096 {
				break
			}
		}

		if response != nil {
			fmt.Println(string(response))
			fmt.Printf("> ")

			//  读取用户输入，发送给target
			reader := bufio.NewReader(os.Stdin)
			buffer, _ := reader.ReadString('\n')
			_, err = nc.conn.Write([]byte(buffer))
			if err != nil {
				fmt.Println(err)
				break
			}

		}
	}
}

func (nc *netcat) listen() {
	cfg := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
			})
		},
	}
	addr := nc.parser.target + ":" + nc.parser.port
	listener, err := cfg.Listen(context.Background(), "tcp", addr)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer listener.Close()
	fmt.Printf("[*] Listening on %s:%s\n", nc.parser.target, nc.parser.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}

		go func(conn net.Conn) {
			fmt.Printf("[*] Accepted connection form %s\n", conn.RemoteAddr().String())
			nc.conn = conn
			nc.handle()
		}(conn)
	}
}

func (nc *netcat) handle() {
	defer nc.conn.Close()

	if nc.parser.execute != "" {
		output := execute(nc.parser.execute)
		_, err := nc.conn.Write([]byte(output))
		if err != nil {
			fmt.Println(err)
		}
	} else if nc.parser.upload != "" {
		// 接收数据
		fileBuffer := []byte{}
		for {
			data := [4096]byte{}
			n, err := nc.conn.Read(data[:])
			if err != nil {
				fmt.Println(err)
				break
			}
			fileBuffer = append(fileBuffer, data[:n]...)
		}

		// 将接收的数据写入文件
		file, err := os.OpenFile(nc.parser.upload, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println(err)
		}

		defer file.Close()

		write := bufio.NewWriter(file)
		_, err = write.WriteString(string(fileBuffer))
		if err != nil {
			fmt.Println(err)
		}

		write.Flush()

		message := fmt.Sprintf("Saved file %s", nc.parser.upload)
		nc.conn.Write([]byte(message))

	} else if nc.parser.command {
		cmdBuffer := []byte{}
		for {
			_, err := nc.conn.Write([]byte("#BHGo: #> "))
			if err != nil {
				fmt.Println(err)
			}

			// 接收数据，直到收到换行符
			for !bytes.Contains(cmdBuffer, []byte{'\n'}) {
				var recv [64]byte
				n, _ := nc.conn.Read(recv[:])
				cmdBuffer = append(cmdBuffer, recv[:n]...)
			}

			// 执行接收到的命令并返回执行结果
			response := execute(string(cmdBuffer))
			if response != "" {
				nc.conn.Write([]byte(response))
			}
			cmdBuffer = nil
		}
	}
}
