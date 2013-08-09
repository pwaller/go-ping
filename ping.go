// Implements pinging of remote hosts via the system ping utility.
//
// Pull requests welcome.
// 
// This implementation is not totally intuitive. It does not invoke ping each
// time a ping request is due to be sent, but instead invokes one ping process
// which is stopped (SIGSTOP) until it is time to send a ping. When it is time
// to send a ping the process is resumed (SIGCONT) until a line of output comes
// from the utility. This means that if multiple ping packet
//
// The current implementation has only been tested in cases where the ping
// utility does not exit unexpectedly. It should panic.
package ping

import (
	"bufio"
	"net/textproto"
	"os"
	"os/exec"
	"syscall"
	"time"
)

// Ping returns a channel of channel which can be used to ping a remote host on
// demand (up to a maximum rate of 1 every 200ms). The ping process is paused
// with a SIGSTOP until a "request" comes in by reading from the  `request`
// channel, which gives the requester a channel over which a line of output from
// the ping utility is put.
// See the implementation of `ping.Pinger` for an example usage and for
// an alternative interface.
func Ping(host string) (request chan chan string) {
	request = make(chan chan string)

	// Ping with the shortest period allowed as non-root (-i0.2 = 200ms)
	// and don't bother with dns lookups (-n)
	// and display a unixtime timestamp at the beginning of the line (-D)
	cmd := exec.Command("ping", "-i0.2", "-n", "-D", host)
	started := false

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	cmd.Stderr = os.Stderr
	if err != nil {
		panic(err)
	}

	go func() {
		r := textproto.NewReader(bufio.NewReader(stdout))

		response := make(chan string)
		for {
			// Wait for a request to transmit the ping
			request <- response

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
					close(request)
				}()

				// Discard first line, usually of the form
				// "PING localhost (127.0.0.1) 56(84) bytes of data."
				_, err := r.ReadLine()
				if err != nil {
					panic(err)
				}
			} else {
				err = cmd.Process.Signal(syscall.SIGCONT)
				if err != nil {
					panic(err)
				}
			}

			line, err := r.ReadLine()
			if err != nil {
				break
			}

			// Immediately stop ping until we're ready to send the next one
			err = cmd.Process.Signal(syscall.SIGSTOP)
			if err != nil {
				panic(err)
			}

			response <- line
		}
	}()

	return
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
