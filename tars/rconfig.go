package tars

import (
	"fmt"
	"os"

	"github.com/TarsCloud/TarsGo/tars/protocol/res/configf"
)

// RConf struct for getting remote config.
type RConf struct {
	app    string
	server string
	comm   *Communicator
	tc     *configf.Config
	path   string
}

// GetRConf returns a default RConf
func GetRConf() *RConf {
	return defaultApp.GetRemoteConf()
}

// GetConfigList get server level config list
func GetConfigList() (fList []string, err error) {
	return defaultApp.GetConfigList()
}

// AddAppConfig add app level config
func AddAppConfig(filename string) (config string, err error) {
	return defaultApp.AddAppConfig(filename)
}

// AddConfig add server level config
func AddConfig(filename string) (config string, err error) {
	return defaultApp.AddConfig(filename)
}

func (a *application) GetRemoteConf() *RConf {
	a.onceRConf.Do(func() {
		cfg := a.ServerConfig()
		a.rConf = NewRConf(cfg.App, cfg.Server, cfg.BasePath)
	})
	return a.rConf
}

// GetConfigList get server level config list
func (a *application) GetConfigList() (fList []string, err error) {
	return a.GetRemoteConf().GetConfigList()
}

// AddAppConfig add app level config
func (a *application) AddAppConfig(filename string) (config string, err error) {
	return a.GetRemoteConf().GetAppConfig(filename)
}

// AddConfig add server level config
func (a *application) AddConfig(filename string) (config string, err error) {
	return a.GetRemoteConf().GetConfig(filename)
}

// NewRConf init a RConf, path should be getting from GetServerConfig().BasePath
func NewRConf(app string, server string, path string) *RConf {
	comm := GetCommunicator()
	obj := comm.app.ServerConfig().Config

	tc := new(configf.Config)
	comm.StringToProxy(obj, tc)
	return &RConf{app, server, comm, tc, path}
}

// GetConfigList is discarded.
func (c *RConf) GetConfigList() (fList []string, err error) {
	info := configf.GetConfigListInfo{
		Appname:    c.app,
		Servername: c.server,
		/*
		   Host:string
		   Setdivision:string
		   Containername:string
		*/
	}
	ret, err := c.tc.ListAllConfigByInfo(&info, &fList, c.comm.Client.Context())
	if err != nil {
		return fList, err
	}
	if ret != 0 {
		return fList, fmt.Errorf("ret:%d", ret)
	}
	return fList, nil
}

// GetAppConfig gets the remote config and save it to the path, also return the content.
func (c *RConf) GetAppConfig(filename string) (config string, err error) {
	info := configf.ConfigInfo{
		Appname:    c.app,
		Servername: "",
		Filename:   filename,
	}
	return c.getConfig(info)
}

// GetConfig gets the remote config and save it to the path, also return the content.
func (c *RConf) GetConfig(filename string) (config string, err error) {
	info := configf.ConfigInfo{
		Appname:    c.app,
		Servername: c.server,
		Filename:   filename,
	}
	return c.getConfig(info)
}

// GetConfig gets the remote config and save it to the path, also return the content.
func (c *RConf) getConfig(info configf.ConfigInfo) (config string, err error) {
	var set string
	if v, ok := c.comm.GetProperty("setdivision"); ok {
		set = v
	}
	info.Setdivision = set
	ret, err := c.tc.LoadConfigByInfo(&info, &config, c.comm.Client.Context())
	if err != nil {
		return config, err
	}
	if ret != 0 {
		return config, fmt.Errorf("ret %d", ret)
	}
	err = saveFile(c.path, info.Filename, config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func saveFile(path string, filename string, content string) error {
	err := os.WriteFile(fmt.Sprintf("%s/%s", path, filename), []byte(content), 0644)
	if err != nil {
		return err
	}
	return nil
}
