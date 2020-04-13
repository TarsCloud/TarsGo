package tars

import (
	"fmt"
	"os"

	"github.com/TarsCloud/TarsGo/tars/util/debug"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

// CheckPanic used to dump stack info to file when catch panic
func CheckPanic() {
	if r := recover(); r != nil {
		var msg string
		if err, ok := r.(error); ok {
			msg = err.Error()
		} else {
			msg = fmt.Sprintf("%#v", r)
		}
		debug.DumpStack(true, "panic", msg)
		rogger.FlushLogger()
		os.Exit(-1)
	}
}
