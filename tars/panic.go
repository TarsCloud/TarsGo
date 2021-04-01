package tars

import (
	"fmt"
	"github.com/TarsCloud/TarsGo/tars/util/debug"
)

// CheckPanic used to dump stack info to file when catch panic
func CheckPanic(onPanics ...func()) {
	if r := recover(); r != nil {
		var msg string
		if err, ok := r.(error); ok {
			msg = err.Error()
		} else {
			msg = fmt.Sprintf("%#v", r)
		}
		debug.DumpStack(true, "panic", msg)

		// onPanic is callback func when catch panic
		for _, onPanic := range onPanics {
			onPanic()
		}
	}
}
