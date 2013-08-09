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
		time.Sleep(500 * time.Millisecond)
		n++
		if n > 1 {
			break
		}
	}
}

func TestPinger(t *testing.T) {
	s := time.Now()
	pinger := Pinger("localhost", 1, 100*time.Millisecond)
	for i := 0; i < 11; i++ {
		println(<-pinger)
	}
	println("Elapsed =", time.Since(s).String())
}
