// +build windows
package grace

type handlerFunc func()

func GraceHandler(stopFunc, userFunc handlerFunc) {
}
