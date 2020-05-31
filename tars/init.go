package tars

//Init need to be called first in some situation:
//like GetServerConfig() and GetClientConfig()
func Init() {
	initOnce.Do(func() {
		initConfig()
	})
}
