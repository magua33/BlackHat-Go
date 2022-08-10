package scanner

import "testing"

func TestScan(t *testing.T) {
	host := "192.168.31.44"
	subnet := "192.168.31.0/24"
	message := "GOGOGO!"
	scan(host, subnet, message)
}
