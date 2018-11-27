// +build linux darwin
package grace

import (
	"os"
	"os/signal"
	"syscall"
)

type handlerFunc func()

func GraceHandler(stopFunc, userFunc handlerFunc) {
	ch := make(chan os.Signal, 10)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGUSR2)
	for {
		sig := <-ch
		switch sig {
		case syscall.SIGKILL:
			signal.Stop(ch)
			stopFunc()
			return
		case syscall.SIGUSR2:
			userFunc()
		}
	}
}
