windows  socket相关的调用请使用WSA*** 或 golang.org/x/sys/windows
windows 需要关闭防火墙(Disable Firewall)



ListenPacket listens for incoming ICMP packets addressed to address. See net.Dial for the syntax of address.
For non-privileged datagram-oriented ICMP endpoints, network must be "udp4" or "udp6".
The endpoint allows to read, write a few limited ICMP messages such as echo request and echo reply. Currently only Darwin and Linux support this.

https://pkg.go.dev/golang.org/x/net/icmp#ListenPacket