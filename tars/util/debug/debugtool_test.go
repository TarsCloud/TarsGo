package debug

import (
	"testing"
)

func TestDumpStack(t *testing.T) {
	DumpStack(true, "testdump")
}
