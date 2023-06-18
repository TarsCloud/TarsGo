package tars

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"

	s "github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/util/trace"
)

// ProxyPrx interface
type ProxyPrx interface {
	SetServant(s.Servant)
}

// Communicator struct
type Communicator struct {
	Client     *clientConfig
	app        *application
	opt        *options
	properties sync.Map
}

// GetCommunicator returns a default communicator
func GetCommunicator() *Communicator {
	return defaultApp.Communicator()
}

// NewCommunicator returns a new communicator. A Communicator is used for communicating with
// the server side which should only init once and be global!!!
func NewCommunicator(opts ...Option) *Communicator {
	return defaultApp.NewCommunicator(opts...)
}

// Communicator returns a default communicator
func (a *application) Communicator() *Communicator {
	a.onceCommunicator.Do(func() {
		a.communicator = a.NewCommunicator()
	})
	return a.communicator
}

func (a *application) NewCommunicator(opts ...Option) *Communicator {
	a.init()
	return newCommunicator(a, a.ClientConfig(), opts...)
}

func newCommunicator(app *application, client *clientConfig, opts ...Option) *Communicator {
	o := *app.opt
	for _, opt := range opts {
		opt(&o)
	}
	c := &Communicator{
		Client: client,
		app:    app,
		opt:    &o,
	}
	c.init()
	return c
}

func (c *Communicator) init() {
	c.SetProperty("locator", c.Client.Locator)
	c.SetProperty("isclient", true)
	c.SetProperty("enableset", false)
	if svrCfg := c.app.ServerConfig(); svrCfg != nil {
		c.SetProperty("notify", svrCfg.Notify)
		c.SetProperty("node", svrCfg.Node)
		c.SetProperty("server", svrCfg.Server)
		c.SetProperty("isclient", false)
		if svrCfg.Enableset {
			c.SetProperty("enableset", true)
			c.SetProperty("setdivision", svrCfg.Setdivision)
		}
	}
}

// GetLocator returns locator as string
func (c *Communicator) GetLocator() string {
	v, _ := c.GetProperty("locator")
	return v
}

// SetLocator sets locator with obj
func (c *Communicator) SetLocator(obj string) {
	c.SetProperty("locator", obj)
	c.Client.Locator = obj
}

// StringToProxy sets the servant of ProxyPrx p with a string servant
func (c *Communicator) StringToProxy(servant string, p ProxyPrx, opts ...EndpointManagerOption) {
	if servant == "" {
		panic("empty servant")
	}
	sp := NewServantProxy(c, servant, opts...)
	p.SetServant(sp)
}

// SetProperty sets communicator property with a string key and an interface value.
// var comm *tars.Communicator
// comm = tars.NewCommunicator()
// e.g. comm.SetProperty("locator", "tars.tarsregistry.QueryObj@tcp -h ... -p ...")
func (c *Communicator) SetProperty(key string, value interface{}) {
	c.properties.Store(key, value)
	c.SetTraceParam(key)
}

// GetProperty returns communicator property value as string and true for key, or empty string
// and false for not exists key
func (c *Communicator) GetProperty(key string) (string, bool) {
	if v, ok := c.properties.Load(key); ok {
		return v.(string), ok
	}
	return "", false
}

func (c *Communicator) SetTraceParam(name string) {
	if len(name) != 0 && name != "trace_param_max_len" {
		return
	}
	defaultValue, ok := c.GetPropertyInt("trace_param_max_len")
	if ok {
		trace.SetTraceParamMaxLen(uint(defaultValue))
	}
}

// GetPropertyInt returns communicator property value as int and true for key, or 0 and false
// for not exists key
func (c *Communicator) GetPropertyInt(key string) (int, bool) {
	if v, ok := c.properties.Load(key); ok {
		return v.(int), ok
	}
	return 0, false
}

// GetPropertyBool returns communicator property value as bool and true for key, or false and false for not exists key
func (c *Communicator) GetPropertyBool(key string) (bool, bool) {
	if v, ok := c.properties.Load(key); ok {
		return v.(bool), ok
	}
	return false, false
}

func (c *Communicator) hashKey() string {
	hash := md5.New()
	hashKeys := []string{"locator", "enableset", "setdivision"}
	for _, k := range hashKeys {
		if v, ok := c.properties.Load(k); ok {
			hash.Write([]byte(fmt.Sprintf("%v:%v", k, v)))
		}
	}
	return hex.EncodeToString(hash.Sum(nil))
}
