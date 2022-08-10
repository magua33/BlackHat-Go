package windows

import "testing"

func TestSniffer(t *testing.T) {
	ADDR := []byte{192, 168, 31, 91}
	sniffer(ADDR)
}
