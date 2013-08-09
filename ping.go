package ping

import (
	"bufio"
	"net/textproto"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func Ping(host string) chan chan string {
	cmd := exec.Command("ping", "-i0.2", host)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Stderr = os.Stderr
	if err != nil {
		panic(err)
	}
	ch := make(chan chan string)

	go func() {
		r := textproto.NewReader(bufio.NewReader(stdout))
		started := false
		response := make(chan string)
		for {
			// Stop ping until we're ready to send one
			// Wait for a request to transmit the ping
			ch <- response

			if !started {
				started = true

				err = cmd.Start()
				if err != nil {
					panic(err)
				}

				go func() {
					err := cmd.Wait()
					if err != nil {
						panic(err)
					}
					close(ch)
				}()

				_, err := r.ReadLine() // Discard first line
				if err != nil {
					panic(err)
				}
			} else {
				err = cmd.Process.Signal(syscall.SIGCONT)
				if err != nil {
					panic(err)
				}
			}

			s, err := r.ReadLine()
			if err != nil {
				break
			}
			err = cmd.Process.Signal(syscall.SIGSTOP)
			if err != nil {
				panic(err)
			}

			response <- s
		}
	}()

	return ch
}

// Ping `host` `n` times per `period`, sending the pings on the returned channel
// Note that as non-root the minimum period is 200ms (see `man 8 ping`)
func Pinger(host string, n int, period time.Duration) chan string {
	ping := Ping(host)

	result := make(chan string)

	go func() {
		for {
			start := time.Now()
			for i := 0; i < n; i++ {
				result <- <-<-ping
			}
			// Sleep however much time remains to make it the correct
			// period
			time.Sleep(period - time.Since(start))
		}
	}()
	return result
}
