package scanner

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

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

type ICMP struct {
	Type int
	Code int
}

func parseICMP(b []byte) ICMP {
	offset := int(b[0]&0xF) * 4
	buf := b[offset : offset+8]

	icmp := ICMP{}
	icmp.Type = int(buf[0])
	icmp.Code = int(buf[1])
	return icmp
}

func scan(host, subnet, message string) {
	go func() {
		time.Sleep(time.Second * 2)
		udpSender(subnet, message)
	}()
	sniff(host, subnet, message)
}

// 给网段内所有可用ip发送udp数据包
func udpSender(subnet, message string) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_IP)
	if err != nil {
		fmt.Println("Init Socket Error:", err)
		return
	}
	defer syscall.Close(fd)

	p := []byte(message)

	ips, _ := hosts(subnet)
	for _, ip := range ips {
		addr, _ := netip.ParseAddr(ip)
		to := syscall.SockaddrInet4{
			Port: 65212,
			Addr: addr.As4(),
		}

		err := syscall.Sendto(fd, p, 0, &to)
		if err != nil {
			fmt.Println("Sendto Error:", err)
		}
	}
}

type void struct{}

var exist void

func sniff(host, subnet, message string) {
	hostUp := make(map[string]struct{})
	hostUp[host+" *"] = exist

	go func() {
		// 创建原始套接字
		fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
		if err != nil {
			fmt.Println("Init Socket Error:", err)
			return
		}
		defer syscall.Close(fd)

		hostByt, _ := netip.ParseAddr(host)

		addr := syscall.SockaddrInet4{
			Port: 0,
			Addr: hostByt.As4(),
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
			rawBuffer := make([]byte, 1024)
			n, _, err := syscall.Recvfrom(fd, rawBuffer, 0)
			if err != nil {
				fmt.Println("Recvfrom Error:", err)
				break
			}

			ipHeader, err := ipv4.ParseHeader(rawBuffer[:n])
			if err != nil {
				fmt.Println("ParseIPv4Header Error:", err)
				break
			}

			if protocol(ipHeader.Protocol) == "ICMP" {
				fmt.Printf("Protocol: %s %v -> %v\n", protocol(ipHeader.Protocol), ipHeader.Src, ipHeader.Dst)
				fmt.Println("Version:", ipHeader.Version)
				fmt.Println("Header Length:", ipHeader.Len, "TTL:", ipHeader.TTL)
				icmpMsg := parseICMP(rawBuffer[:n])

				fmt.Println("ICMP -> Type:", icmpMsg.Type, "Code:", icmpMsg.Code)

				if icmpMsg.Code == 3 && icmpMsg.Type == 3 {
					network, err := netip.ParsePrefix(subnet)
					if err != nil {
						fmt.Println("ParsePrefix Error:", err)
						continue
					}

					ip, err := netip.ParseAddr(ipHeader.Src.String())
					if err != nil {
						fmt.Println("ParseAddr Error:", err)
						continue
					}

					// 检查ip是否在网段内
					fmt.Println(ip, network, network.Contains(ip))
					fmt.Println(rawBuffer[:n])
					if network.Contains(ip) {
						if string(rawBuffer[n-len([]byte(message)):n]) == message {
							tgt := ipHeader.Src.String()
							_, inHostUp := hostUp[tgt]
							if tgt != host && !inHostUp {
								hostUp[tgt] = exist
							}
						}
					}
				}
			}
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Printf("\nUser interrupted.")
	if hostUp != nil {
		fmt.Printf("\n\nSummary Hosts up on %s\n", subnet)
	}

	for host := range hostUp {
		fmt.Println(host)
	}
}

// hosts 获取网段内所有可用ip
func hosts(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	lenIPs := len(ips)
	switch {
	case lenIPs < 2:
		return ips, nil

	default:
		return ips[1 : len(ips)-1], nil
	}
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
