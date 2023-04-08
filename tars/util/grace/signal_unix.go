//go:build linux || darwin
// +build linux darwin

package grace

import (
	"os"
	"os/signal"
	"syscall"
)

type handlerFunc func()

// GraceHandler set the signal handler for grace restart
func GraceHandler(userFunc, stopFunc handlerFunc) {
	ch := make(chan os.Signal, 10)
	// remove syscall.SIGKILL
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGTERM, os.Interrupt)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGUSR1:
			userFunc()
		case syscall.SIGUSR2, syscall.SIGTERM, os.Interrupt: // remove syscall.SIGKILL
			signal.Stop(ch)
			stopFunc()
		}
	}
}

// SignalUSR2 send signal USR2 to pid
func SignalUSR2(pid int) {
	syscall.Kill(pid, syscall.SIGUSR2)
}
