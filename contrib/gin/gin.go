package gin

import (
	"net/http"

	"github.com/TarsCloud/TarsGo/tars"
	"github.com/gin-gonic/gin"
)

type server struct {
	*gin.Engine
	cfg *tars.TarsHttpConf
}

type ServerOption func(*server)

var _ tars.HttpHandler = (*server)(nil)

func New(opts ...ServerOption) tars.HttpHandler {
	s := &server{
		Engine: gin.Default(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (g *server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	g.Engine.ServeHTTP(w, req)
}

func (g *server) SetConfig(cfg *tars.TarsHttpConf) {
	g.cfg = cfg
}
