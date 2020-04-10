// +build go1.13 go1.14

package tars

import (
	"testing"
)

//Init need to be called first in some situation:
//like GetServerConfig() and GetClientConfig()
//In Go1.13 should call testing.Init() before call init()
func Init() {
	initOnce.Do(func() {
		testing.Init()
		initConfig()
	})
}
