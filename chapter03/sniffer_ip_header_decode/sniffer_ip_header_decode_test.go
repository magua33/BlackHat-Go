package sniffer_ip_header_decode

import "testing"

func TestSniff(t *testing.T) {
	host := [4]byte{192, 168, 31, 44}
	sniff(host)
}
