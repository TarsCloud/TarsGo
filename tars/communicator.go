package tars

import (
	s "github.com/TarsCloud/TarsGo/tars/model"
	"sync"
)

type ProxyPrx interface {
	SetServant(s.Servant)
}

func Dail(servant string) *ServantProxy {
	c := new(Communicator)
	c.init()
	return c.s.GetServantProxy(servant)
}

func NewCommunicator() *Communicator {
	c := new(Communicator)
	c.init()
	return c
}

type Communicator struct {
	s          *ServantProxyFactory
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
			"",
			"",
			"",
			"",
			refreshEndpointInterval,
			reportInterval,
			AsyncInvokeTimeout,
		}
	}
	c.SetProperty("netthread", 2)
	c.SetProperty("isclient", true)
	c.SetProperty("enableset", false)
	if GetServerConfig() != nil {
		c.SetProperty("netthread", GetServerConfig().netThread)
		c.SetProperty("notify", GetServerConfig().notify)
		c.SetProperty("node", GetServerConfig().Node)
		c.SetProperty("server", GetServerConfig().Server)
		c.SetProperty("isclient", false)
		if GetServerConfig().Enableset {
			c.SetProperty("enableset", true)
			c.SetProperty("setdivision", GetServerConfig().Setdivision)
		}
	}

	c.s = new(ServantProxyFactory)
	c.s.Init(c)
}
func (c *Communicator) GetLocator() string {
	v, _ := c.GetProperty("locator")
	return v
}
func (c *Communicator) SetLocator(obj string) {
	c.SetProperty("locator", obj)
}
func (c *Communicator) StringToProxy(servant string, p ProxyPrx) {
	p.SetServant(c.s.GetServantProxy(servant))
}

func (c *Communicator) SetProperty(key string, value interface{}) {
	c.properties.Store(key, value)
}
func (c *Communicator) GetProperty(key string) (string, bool) {
	if v, ok := c.properties.Load(key); ok {
		return v.(string), ok
	}
	return "", false
}
func (c *Communicator) GetPropertyInt(key string) (int, bool) {
	if v, ok := c.properties.Load(key); ok {
		return v.(int), ok
	}
	return 0, false
}
func (c *Communicator) GetPropertyBool(key string) (bool, bool) {
	if v, ok := c.properties.Load(key); ok {
		return v.(bool), ok
	}
	return false, false
}
