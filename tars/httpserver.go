package tars

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/statf"
)

var realIPHeader []string

func init() {
	realIPHeader = []string{ // the order is important
		"X-Real-Ip",
		"X-Forwarded-For-Pound",
		"X-Forwarded-For",
	}
}

// TarsHttpConf is configuration for tars http server.
type TarsHttpConf struct {
	Container              string
	AppName                string
	IP                     string
	Port                   int32
	Version                string
	SetId                  string
	ExceptionStatusChecker func(int) bool
}

// TarsHttpMux is http.ServeMux for tars http server.
type TarsHttpMux struct {
	http.ServeMux
	cfg *TarsHttpConf
}

type httpStatInfo struct {
	reqAddr    string
	pattern    string
	statusCode int
	costTime   int64
}

// ServeHTTP is the server for the TarsHttpMux.
func (mux *TarsHttpMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, pattern := mux.Handler(r)
	tw := &TarsResponseWriter{w, 0}
	startTime := time.Now().UnixNano() / 1e6
	h.ServeHTTP(tw, r)
	costTime := int64(time.Now().UnixNano()/1e6 - startTime)
	var reqAddr string
	for _, h := range realIPHeader {
		reqAddr = r.Header.Get(h)
		if reqAddr != "" {
			break
		}
	}
	if reqAddr == "" { // no proxy
		reqAddr = strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}
	if pattern == "" {
		pattern = "/"
	}
	st := &httpStatInfo{
		reqAddr:    reqAddr,
		pattern:    pattern,
		statusCode: tw.StatusCode,
		costTime:   costTime,
	}
	go mux.reportHttpStat(st)
}

func (mux *TarsHttpMux) reportHttpStat(st *httpStatInfo) {
	if mux.cfg == nil || StatReport == nil {
		return
	}
	cfg := mux.cfg
	var _statInfo = statf.StatMicMsgHead{}
	_statInfo.MasterName = "http_client"
	_statInfo.MasterIp = st.reqAddr

	_statInfo.TarsVersion = cfg.Version
	_statInfo.SlaveName = cfg.AppName
	_statInfo.SlaveIp = cfg.IP // from server
	_statInfo.SlavePort = cfg.Port
	//_statInfo.SSlaveContainer = cfg.Container
	_statInfo.InterfaceName = st.pattern
	if cfg.SetId != "" {
		setList := strings.Split(cfg.SetId, ".")
		_statInfo.SlaveSetName = setList[0]
		_statInfo.SlaveSetArea = setList[1]
		_statInfo.SlaveSetID = setList[2]
		//被调也要加上set信息
		_statInfo.SlaveName = fmt.Sprintf("%s.%s%s%s", _statInfo.SlaveName, setList[0], setList[1], setList[2])
	}

	var _statBody = statf.StatMicMsgBody{}
	exceptionChecker := mux.cfg.ExceptionStatusChecker
	if exceptionChecker == nil {
		// if nil, use default
		exceptionChecker = DefaultExceptionStatusChecker
	}
	if exceptionChecker(st.statusCode) {
		_statBody.ExecCount = 1 // 异常
	} else {
		_statBody.Count = 1
		_statBody.TotalRspTime = st.costTime
		_statBody.MaxRspTime = int32(st.costTime)
		_statBody.MinRspTime = int32(st.costTime)
	}

	info := StatInfo{}
	info.Head = _statInfo
	info.Body = _statBody
	StatReport.pushBackMsg(info, true)
}

// SetConfig sets the cfg tho the TarsHttpMux.
func (mux *TarsHttpMux) SetConfig(cfg *TarsHttpConf) {
	mux.cfg = cfg
}

// DefaultExceptionStatusChecker Default Exception Status Checker
func DefaultExceptionStatusChecker(statusCode int) bool {
	return statusCode >= 400
}

// TarsResponseWriter is http.ResponseWriter for tars.
type TarsResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader is used for write the http header with the http code.
func (w *TarsResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// Hijack add Hijack method for TarsResponseWriter
func (w *TarsResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, errors.New("http.Hijacker is unavailable on the writer")
}
