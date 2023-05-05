package tars

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/TarsCloud/TarsGo/tars/util/debug"
	"github.com/TarsCloud/TarsGo/tars/util/rogger"
)

// Admin struct
type Admin struct {
	app *application
}

type adminFn func(string) (string, error)

func newAdmin(app *application) *Admin {
	return &Admin{app: app}
}

// RegisterAdmin register admin functions
func RegisterAdmin(name string, fn adminFn) {
	defaultApp.RegisterAdmin(name, fn)
}

// RegisterAdmin register admin functions
func (a *application) RegisterAdmin(name string, fn adminFn) {
	a.adminMethods[name] = fn
}

// Shutdown all servant by admin
func (a *Admin) Shutdown() error {
	atomic.StoreInt32(&a.app.isShutdownByAdmin, 1)
	go a.app.graceShutdown()
	return nil
}

// Notify handler for cmds from admin
func (a *Admin) Notify(command string) (string, error) {
	cmd := strings.Split(command, " ")
	// report command to notify
	go ReportNotifyInfo(NotifyNormal, "AdminServant::notify:"+command)
	switch cmd[0] {
	case "tars.viewversion":
		return a.app.ServerConfig().Version, nil
	case "tars.setloglevel":
		if len(cmd) >= 2 {
			a.app.appCache.LogLevel = cmd[1]
			switch cmd[1] {
			case "INFO":
				rogger.SetLevel(rogger.INFO)
			case "WARN":
				rogger.SetLevel(rogger.WARN)
			case "ERROR":
				rogger.SetLevel(rogger.ERROR)
			case "DEBUG":
				rogger.SetLevel(rogger.DEBUG)
			case "NONE":
				rogger.SetLevel(rogger.OFF)
			default:
				return fmt.Sprintf("%s failed: unknown log level [%s]!", cmd[0], cmd[1]), nil
			}
			return fmt.Sprintf("%s succ", command), nil
		}
		return fmt.Sprintf("%s failed: missing loglevel!", command), nil
	case "tars.dumpstack":
		debug.DumpStack(true, "stackinfo", "tars.dumpstack:")
		return fmt.Sprintf("%s succ", command), nil
	case "tars.loadconfig":
		cfg := a.app.ServerConfig()
		remoteConf := NewRConf(cfg.App, cfg.Server, cfg.BasePath)
		_, err := remoteConf.GetConfig(cmd[1])
		if err != nil {
			return fmt.Sprintf("Getconfig Error!: %s", cmd[1]), err
		}
		return fmt.Sprintf("Getconfig Success!: %s", cmd[1]), nil
	case "tars.connection":
		return fmt.Sprintf("%s not support now!", command), nil
	case "tars.gracerestart":
		a.app.graceRestart()
		return "restart gracefully!", nil
	case "tars.pprof":
		port := ":8080"
		timeout := time.Second * 600
		if len(cmd) > 1 {
			port = ":" + cmd[1]
		}
		if len(cmd) > 2 {
			t, _ := strconv.ParseInt(cmd[2], 10, 64)
			if 0 < t && t < 3600 {
				timeout = time.Second * time.Duration(t)
			}
		}
		cfg := a.app.ServerConfig()
		addr := cfg.LocalIP + port
		go func() {
			mux := http.NewServeMux()
			mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
			mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
			mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
			mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
			mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
			s := &http.Server{Addr: addr, Handler: mux}
			TLOG.Info("start serve pprof ", addr)
			go s.ListenAndServe()
			time.Sleep(timeout)
			s.Shutdown(context.Background())
			TLOG.Info("stop serve pprof ", addr)
		}()
		return "see http://" + addr + "/debug/pprof/", nil
	default:
		if fn, ok := a.app.adminMethods[cmd[0]]; ok {
			return fn(command)
		}
		return fmt.Sprintf("%s not support now!", command), nil
	}
}
