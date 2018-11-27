// +build linux darwin
package grace

import (
	"os"
	"os/signal"
	"syscall"
)

type handlerFunc func()

func GraceHandler(userFunc, stopFunc handlerFunc) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGKILL, syscall.SIGTERM)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGUSR1:
			userFunc()
		case syscall.SIGUSR2, syscall.SIGKILL, syscall.SIGTERM:
			signal.Stop(ch)
			stopFunc()
		}
	}
}
