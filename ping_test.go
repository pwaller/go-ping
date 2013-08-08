package ping

import (
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	ping := Ping("8.8.8.8")
	n := 0
	for r := range ping {
		t := time.Now()
		println("t = ", <-r)
		resp := time.Since(t)
		println(" -- ", resp.String())
		time.Sleep(2 * time.Second)
		n++
		if n > 1 {
			break
		}
	}
}
