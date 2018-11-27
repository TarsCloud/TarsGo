// +build windows
package grace

type handlerFunc func()

// GraceHandler is now supported in windows
func GraceHandler(stopFunc, userFunc handlerFunc) {
}
