package ping

import (
	"bufio"
	"net/textproto"
	"os"
	"os/exec"
	"syscall"
)

func Ping(host string) chan chan string {
	cmd := exec.Command("ping", host)

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
