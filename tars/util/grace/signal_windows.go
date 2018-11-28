// +build windows
package grace

type handlerFunc func()

// GraceHandler is not supported in windows
func GraceHandler(stopFunc, userFunc handlerFunc) {
}
