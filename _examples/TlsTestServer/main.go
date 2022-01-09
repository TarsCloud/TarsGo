package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/TarsCloud/TarsGo/tars"

	"TlsTestServer/tars-protocol/App"
)

func main() {
	tars.ServerConfigPath = "config.conf"
	// Get server config
	cfg := tars.GetServerConfig()

	// New servant imp
	imp := new(TlsImp)
	err := imp.Init()
	if err != nil {
		fmt.Printf("TlsImp init fail, err:(%s)\n", err)
		os.Exit(-1)
	}
	// New servant
	app := new(App.Tls)
	// Register Servant
	app.AddServantWithContext(imp, cfg.App+"."+cfg.Server+".TlsObj")
	mux := &tars.TarsHttpMux{}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano()/1e6)))
	})
	//Register http server
	tars.AddHttpServant(mux, cfg.App+"."+cfg.Server+".HttpsObj")

	// Run application
	tars.Run()
}
