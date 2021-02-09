package tars

import "fmt"

// Error is the type of rpc error with error code
type Error struct {
	Code    int32
	Message string
}

// Error returns the error message
func (e *Error) Error() string {
	return e.Message
}

// GetErrorCode returns the error code
func GetErrorCode(err error) int32 {
	if err == nil {
		return 0
	}
	e, ok := err.(*Error)
	if !ok {
		return 1
	}
	return e.Code
}

// Errorf return the tars.Error instance
func Errorf(code int32, format string, args ...interface{}) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}
