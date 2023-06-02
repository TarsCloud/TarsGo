package gin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/statf"
	"github.com/gin-gonic/gin"
)

type ServerOption func(*Server)

type Server struct {
	*gin.Engine
	cfg *tars.TarsHttpConf
}

var _ tars.HttpHandler = (*Server)(nil)

func New(opts ...ServerOption) *Server {
	s := &Server{
		Engine: gin.Default(),
	}
	for _, opt := range opts {
		opt(s)
	}
	s.middleware()
	return s
}

func (g *Server) middleware() {
	g.Use(func(c *gin.Context) {
		startTime := time.Now()
		if c.Request.RequestURI == "*" {
			if c.Request.ProtoAtLeast(1, 1) {
				c.Header("Connection", "close")
			}
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Next()
		costTime := time.Since(startTime).Milliseconds()
		go g.reportHttpStat(c.ClientIP(), c.FullPath(), c.Writer.Status(), costTime)
	})
}

func (g *Server) SetConfig(cfg *tars.TarsHttpConf) {
	g.cfg = cfg
}

func (g *Server) reportHttpStat(clientIP, pattern string, statusCode int, costTime int64) {
	if g.cfg == nil {
		return
	}
	cfg := g.cfg
	var statInfo = statf.StatMicMsgHead{}
	statInfo.MasterName = "http_client"
	statInfo.MasterIp = clientIP

	statInfo.TarsVersion = cfg.Version
	statInfo.SlaveName = cfg.AppName
	statInfo.SlaveIp = cfg.IP // from server
	statInfo.SlavePort = cfg.Port
	statInfo.InterfaceName = pattern
	if cfg.SetId != "" {
		setList := strings.Split(cfg.SetId, ".")
		statInfo.SlaveSetName = setList[0]
		statInfo.SlaveSetArea = setList[1]
		statInfo.SlaveSetID = setList[2]
		//被调也要加上set信息
		statInfo.SlaveName = fmt.Sprintf("%s.%s%s%s", statInfo.SlaveName, setList[0], setList[1], setList[2])
	}

	var statBody = statf.StatMicMsgBody{}
	exceptionChecker := g.cfg.ExceptionStatusChecker
	if exceptionChecker == nil {
		// if nil, use default
		exceptionChecker = tars.DefaultExceptionStatusChecker
	}
	if exceptionChecker(statusCode) {
		statBody.ExecCount = 1 // 异常
	} else {
		statBody.Count = 1
		statBody.TotalRspTime = costTime
		statBody.MaxRspTime = int32(costTime)
		statBody.MinRspTime = int32(costTime)
	}

	tars.ReportStatBase(&statInfo, &statBody, true)
}
