package tars

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"

	s "github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

// ProxyPrx interface
type ProxyPrx interface {
	SetServant(s.Servant)
}

// NewCommunicator returns a new communicator. A Communicator is used for communicating with
// the server side which should only init once and be global!!!
func NewCommunicator() *Communicator {
	c := new(Communicator)
	c.init()
	return c
}

// Communicator struct
type Communicator struct {
	Client     *clientConfig
	properties sync.Map
}

func (c *Communicator) init() {
	if GetClientConfig() != nil {
		c.SetProperty("locator", GetClientConfig().Locator)
		//TODO
		c.Client = GetClientConfig()
	} else {
		c.Client = &clientConfig{
			RefreshEndpointInterval: refreshEndpointInterval,
			ReportInterval:          reportInterval,
			CheckStatusInterval:     checkStatusInterval,
			AsyncInvokeTimeout:      AsyncInvokeTimeout,
			ClientQueueLen:          ClientQueueLen,
			ClientIdleTimeout:       tools.ParseTimeOut(ClientIdleTimeout),
			ClientReadTimeout:       tools.ParseTimeOut(ClientReadTimeout),
			ClientWriteTimeout:      tools.ParseTimeOut(ClientWriteTimeout),
			ClientDialTimeout:       tools.ParseTimeOut(ClientDialTimeout),
			ReqDefaultTimeout:       ReqDefaultTimeout,
			ObjQueueMax:             ObjQueueMax,
			AdapterProxyTicker:      tools.ParseTimeOut(AdapterProxyTicker),
			AdapterProxyResetCount:  AdapterProxyResetCount,
		}
	}
	c.SetProperty("isclient", true)
	c.SetProperty("enableset", false)
	if GetServerConfig() != nil {
		c.SetProperty("notify", GetServerConfig().Notify)
		c.SetProperty("node", GetServerConfig().Node)
		c.SetProperty("server", GetServerConfig().Server)
		c.SetProperty("isclient", false)
		if GetServerConfig().Enableset {
			c.SetProperty("enableset", true)
			c.SetProperty("setdivision", GetServerConfig().Setdivision)
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
}

// StringToProxy sets the servant of ProxyPrx p with a string servant
func (c *Communicator) StringToProxy(servant string, p ProxyPrx) {
	if servant == "" {
		panic("empty servant")
	}
	sp := newServantProxy(c, servant)
	p.SetServant(sp)
}

// SetProperty sets communicator property with a string key and an interface value.
// var comm *tars.Communicator
// comm = tars.NewCommunicator()
// e.g. comm.SetProperty("locator", "tars.tarsregistry.QueryObj@tcp -h ... -p ...")
func (c *Communicator) SetProperty(key string, value interface{}) {
	c.properties.Store(key, value)
}

// GetProperty returns communicator property value as string and true for key, or empty string
// and false for not exists key
func (c *Communicator) GetProperty(key string) (string, bool) {
	if v, ok := c.properties.Load(key); ok {
		return v.(string), ok
	}
	return "", false
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
