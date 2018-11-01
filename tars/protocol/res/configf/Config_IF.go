//Package configf comment
// This file war generated by tars2go 1.1
// Generated from Config.tars
package configf

import (
	"context"
	"fmt"
	m "github.com/TarsCloud/TarsGo/tars/model"
	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
	"github.com/TarsCloud/TarsGo/tars/protocol/res/requestf"
	"github.com/TarsCloud/TarsGo/tars/util/tools"
)

//Config struct
type Config struct {
	s m.Servant
}

//ListConfig is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) ListConfig(App string, Server string, Vf *[]string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(App, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Server, 2)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "ListConfig", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err, have, ty = _is.SkipToNoCheck(3, true)
	if err != nil {
		return ret, err
	}

	if ty == codec.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return ret, err
		}
		(*Vf) = make([]string, length, length)
		for i0, e0 := int32(0), length; i0 < e0; i0++ {

			err = _is.Read_string(&(*Vf)[i0], 0, false)
			if err != nil {
				return ret, err
			}
		}
	} else if ty == codec.SIMPLE_LIST {
		err = fmt.Errorf("not support simple_list type")
		if err != nil {
			return ret, err
		}
	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return ret, err
		}
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ListConfigWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) ListConfigWithContext(ctx context.Context, App string, Server string, Vf *[]string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(App, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Server, 2)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "ListConfig", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err, have, ty = _is.SkipToNoCheck(3, true)
	if err != nil {
		return ret, err
	}

	if ty == codec.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return ret, err
		}
		(*Vf) = make([]string, length, length)
		for i1, e1 := int32(0), length; i1 < e1; i1++ {

			err = _is.Read_string(&(*Vf)[i1], 0, false)
			if err != nil {
				return ret, err
			}
		}
	} else if ty == codec.SIMPLE_LIST {
		err = fmt.Errorf("not support simple_list type")
		if err != nil {
			return ret, err
		}
	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return ret, err
		}
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//LoadConfig is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) LoadConfig(App string, Server string, Filename string, Config *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(App, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Server, 2)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Filename, 3)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "loadConfig", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Config), 4, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//LoadConfigWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) LoadConfigWithContext(ctx context.Context, App string, Server string, Filename string, Config *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(App, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Server, 2)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Filename, 3)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "loadConfig", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Config), 4, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//LoadConfigByHost is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) LoadConfigByHost(AppServerName string, Filename string, Host string, Config *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(AppServerName, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Filename, 2)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Host, 3)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "loadConfigByHost", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Config), 4, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//LoadConfigByHostWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) LoadConfigByHostWithContext(ctx context.Context, AppServerName string, Filename string, Host string, Config *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(AppServerName, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Filename, 2)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Host, 3)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "loadConfigByHost", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Config), 4, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//CheckConfig is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) CheckConfig(AppServerName string, Filename string, Host string, Result *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(AppServerName, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Filename, 2)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Host, 3)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "checkConfig", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Result), 4, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//CheckConfigWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) CheckConfigWithContext(ctx context.Context, AppServerName string, Filename string, Host string, Result *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = _os.Write_string(AppServerName, 1)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Filename, 2)
	if err != nil {
		return ret, err
	}

	err = _os.Write_string(Host, 3)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "checkConfig", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Result), 4, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ListConfigByInfo is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) ListConfigByInfo(ConfigInfo *ConfigInfo, Vf *[]string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "ListConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err, have, ty = _is.SkipToNoCheck(2, true)
	if err != nil {
		return ret, err
	}

	if ty == codec.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return ret, err
		}
		(*Vf) = make([]string, length, length)
		for i2, e2 := int32(0), length; i2 < e2; i2++ {

			err = _is.Read_string(&(*Vf)[i2], 0, false)
			if err != nil {
				return ret, err
			}
		}
	} else if ty == codec.SIMPLE_LIST {
		err = fmt.Errorf("not support simple_list type")
		if err != nil {
			return ret, err
		}
	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return ret, err
		}
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ListConfigByInfoWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) ListConfigByInfoWithContext(ctx context.Context, ConfigInfo *ConfigInfo, Vf *[]string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "ListConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err, have, ty = _is.SkipToNoCheck(2, true)
	if err != nil {
		return ret, err
	}

	if ty == codec.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return ret, err
		}
		(*Vf) = make([]string, length, length)
		for i3, e3 := int32(0), length; i3 < e3; i3++ {

			err = _is.Read_string(&(*Vf)[i3], 0, false)
			if err != nil {
				return ret, err
			}
		}
	} else if ty == codec.SIMPLE_LIST {
		err = fmt.Errorf("not support simple_list type")
		if err != nil {
			return ret, err
		}
	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return ret, err
		}
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//LoadConfigByInfo is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) LoadConfigByInfo(ConfigInfo *ConfigInfo, Config *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "loadConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Config), 2, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//LoadConfigByInfoWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) LoadConfigByInfoWithContext(ctx context.Context, ConfigInfo *ConfigInfo, Config *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "loadConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Config), 2, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//CheckConfigByInfo is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) CheckConfigByInfo(ConfigInfo *ConfigInfo, Result *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "checkConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Result), 2, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//CheckConfigByInfoWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) CheckConfigByInfoWithContext(ctx context.Context, ConfigInfo *ConfigInfo, Result *string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "checkConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err = _is.Read_string(&(*Result), 2, true)
	if err != nil {
		return ret, err
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ListAllConfigByInfo is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) ListAllConfigByInfo(ConfigInfo *GetConfigListInfo, Vf *[]string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	ctx := context.Background()
	err = _obj.s.Tars_invoke(ctx, 0, "ListAllConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err, have, ty = _is.SkipToNoCheck(2, true)
	if err != nil {
		return ret, err
	}

	if ty == codec.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return ret, err
		}
		(*Vf) = make([]string, length, length)
		for i4, e4 := int32(0), length; i4 < e4; i4++ {

			err = _is.Read_string(&(*Vf)[i4], 0, false)
			if err != nil {
				return ret, err
			}
		}
	} else if ty == codec.SIMPLE_LIST {
		err = fmt.Errorf("not support simple_list type")
		if err != nil {
			return ret, err
		}
	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return ret, err
		}
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//ListAllConfigByInfoWithContext is the proxy function for the method defined in the tars file, with the context
func (_obj *Config) ListAllConfigByInfoWithContext(ctx context.Context, ConfigInfo *GetConfigListInfo, Vf *[]string, _opt ...map[string]string) (ret int32, err error) {

	var length int32
	var have bool
	var ty byte
	_os := codec.NewBuffer()
	err = ConfigInfo.WriteBlock(_os, 1)
	if err != nil {
		return ret, err
	}

	var _status map[string]string
	var _context map[string]string
	_resp := new(requestf.ResponsePacket)
	err = _obj.s.Tars_invoke(ctx, 0, "ListAllConfigByInfo", _os.ToBytes(), _status, _context, _resp)
	if err != nil {
		return ret, err
	}
	_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))
	err = _is.Read_int32(&ret, 0, true)
	if err != nil {
		return ret, err
	}

	err, have, ty = _is.SkipToNoCheck(2, true)
	if err != nil {
		return ret, err
	}

	if ty == codec.LIST {
		err = _is.Read_int32(&length, 0, true)
		if err != nil {
			return ret, err
		}
		(*Vf) = make([]string, length, length)
		for i5, e5 := int32(0), length; i5 < e5; i5++ {

			err = _is.Read_string(&(*Vf)[i5], 0, false)
			if err != nil {
				return ret, err
			}
		}
	} else if ty == codec.SIMPLE_LIST {
		err = fmt.Errorf("not support simple_list type")
		if err != nil {
			return ret, err
		}
	} else {
		err = fmt.Errorf("require vector, but not")
		if err != nil {
			return ret, err
		}
	}

	_ = length
	_ = have
	_ = ty
	return ret, nil
}

//SetServant sets servant for the service.
func (_obj *Config) SetServant(s m.Servant) {
	_obj.s = s
}

//TarsSetTimeout sets the timeout for the servant which is in ms.
func (_obj *Config) TarsSetTimeout(t int) {
	_obj.s.TarsSetTimeout(t)
}

type _impConfig interface {
	ListConfig(App string, Server string, Vf *[]string) (ret int32, err error)
	LoadConfig(App string, Server string, Filename string, Config *string) (ret int32, err error)
	LoadConfigByHost(AppServerName string, Filename string, Host string, Config *string) (ret int32, err error)
	CheckConfig(AppServerName string, Filename string, Host string, Result *string) (ret int32, err error)
	ListConfigByInfo(ConfigInfo *ConfigInfo, Vf *[]string) (ret int32, err error)
	LoadConfigByInfo(ConfigInfo *ConfigInfo, Config *string) (ret int32, err error)
	CheckConfigByInfo(ConfigInfo *ConfigInfo, Result *string) (ret int32, err error)
	ListAllConfigByInfo(ConfigInfo *GetConfigListInfo, Vf *[]string) (ret int32, err error)
}
type _impConfigWithContext interface {
	ListConfig(ctx context.Context, App string, Server string, Vf *[]string) (ret int32, err error)
	LoadConfig(ctx context.Context, App string, Server string, Filename string, Config *string) (ret int32, err error)
	LoadConfigByHost(ctx context.Context, AppServerName string, Filename string, Host string, Config *string) (ret int32, err error)
	CheckConfig(ctx context.Context, AppServerName string, Filename string, Host string, Result *string) (ret int32, err error)
	ListConfigByInfo(ctx context.Context, ConfigInfo *ConfigInfo, Vf *[]string) (ret int32, err error)
	LoadConfigByInfo(ctx context.Context, ConfigInfo *ConfigInfo, Config *string) (ret int32, err error)
	CheckConfigByInfo(ctx context.Context, ConfigInfo *ConfigInfo, Result *string) (ret int32, err error)
	ListAllConfigByInfo(ctx context.Context, ConfigInfo *GetConfigListInfo, Vf *[]string) (ret int32, err error)
}

//Dispatch is used to call the server side implemnet for the method defined in the tars file. withContext shows using context or not.
func (_obj *Config) Dispatch(ctx context.Context, _val interface{}, req *requestf.RequestPacket, resp *requestf.ResponsePacket, withContext bool) (err error) {
	var length int32
	var have bool
	var ty byte
	_is := codec.NewReader(tools.Int8ToByte(req.SBuffer))
	_os := codec.NewBuffer()
	switch req.SFuncName {
	case "ListConfig":
		var App string
		err = _is.Read_string(&App, 1, true)
		if err != nil {
			return err
		}
		var Server string
		err = _is.Read_string(&Server, 2, true)
		if err != nil {
			return err
		}
		var Vf []string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.ListConfig(App, Server, &Vf)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.ListConfig(ctx, App, Server, &Vf)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.WriteHead(codec.LIST, 3)
		if err != nil {
			return err
		}
		err = _os.Write_int32(int32(len(Vf)), 0)
		if err != nil {
			return err
		}
		for _, v := range Vf {

			err = _os.Write_string(v, 0)
			if err != nil {
				return err
			}
		}
	case "loadConfig":
		var App string
		err = _is.Read_string(&App, 1, true)
		if err != nil {
			return err
		}
		var Server string
		err = _is.Read_string(&Server, 2, true)
		if err != nil {
			return err
		}
		var Filename string
		err = _is.Read_string(&Filename, 3, true)
		if err != nil {
			return err
		}
		var Config string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.LoadConfig(App, Server, Filename, &Config)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.LoadConfig(ctx, App, Server, Filename, &Config)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.Write_string(Config, 4)
		if err != nil {
			return err
		}
	case "loadConfigByHost":
		var AppServerName string
		err = _is.Read_string(&AppServerName, 1, true)
		if err != nil {
			return err
		}
		var Filename string
		err = _is.Read_string(&Filename, 2, true)
		if err != nil {
			return err
		}
		var Host string
		err = _is.Read_string(&Host, 3, true)
		if err != nil {
			return err
		}
		var Config string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.LoadConfigByHost(AppServerName, Filename, Host, &Config)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.LoadConfigByHost(ctx, AppServerName, Filename, Host, &Config)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.Write_string(Config, 4)
		if err != nil {
			return err
		}
	case "checkConfig":
		var AppServerName string
		err = _is.Read_string(&AppServerName, 1, true)
		if err != nil {
			return err
		}
		var Filename string
		err = _is.Read_string(&Filename, 2, true)
		if err != nil {
			return err
		}
		var Host string
		err = _is.Read_string(&Host, 3, true)
		if err != nil {
			return err
		}
		var Result string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.CheckConfig(AppServerName, Filename, Host, &Result)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.CheckConfig(ctx, AppServerName, Filename, Host, &Result)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.Write_string(Result, 4)
		if err != nil {
			return err
		}
	case "ListConfigByInfo":
		var ConfigInfo ConfigInfo
		err = ConfigInfo.ReadBlock(_is, 1, true)
		if err != nil {
			return err
		}
		var Vf []string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.ListConfigByInfo(&ConfigInfo, &Vf)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.ListConfigByInfo(ctx, &ConfigInfo, &Vf)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.WriteHead(codec.LIST, 2)
		if err != nil {
			return err
		}
		err = _os.Write_int32(int32(len(Vf)), 0)
		if err != nil {
			return err
		}
		for _, v := range Vf {

			err = _os.Write_string(v, 0)
			if err != nil {
				return err
			}
		}
	case "loadConfigByInfo":
		var ConfigInfo ConfigInfo
		err = ConfigInfo.ReadBlock(_is, 1, true)
		if err != nil {
			return err
		}
		var Config string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.LoadConfigByInfo(&ConfigInfo, &Config)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.LoadConfigByInfo(ctx, &ConfigInfo, &Config)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.Write_string(Config, 2)
		if err != nil {
			return err
		}
	case "checkConfigByInfo":
		var ConfigInfo ConfigInfo
		err = ConfigInfo.ReadBlock(_is, 1, true)
		if err != nil {
			return err
		}
		var Result string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.CheckConfigByInfo(&ConfigInfo, &Result)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.CheckConfigByInfo(ctx, &ConfigInfo, &Result)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.Write_string(Result, 2)
		if err != nil {
			return err
		}
	case "ListAllConfigByInfo":
		var ConfigInfo GetConfigListInfo
		err = ConfigInfo.ReadBlock(_is, 1, true)
		if err != nil {
			return err
		}
		var Vf []string
		if withContext == false {
			_imp := _val.(_impConfig)
			ret, err := _imp.ListAllConfigByInfo(&ConfigInfo, &Vf)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		} else {
			_imp := _val.(_impConfigWithContext)
			ret, err := _imp.ListAllConfigByInfo(ctx, &ConfigInfo, &Vf)
			if err != nil {
				return err
			}

			err = _os.Write_int32(ret, 0)
			if err != nil {
				return err
			}
		}

		err = _os.WriteHead(codec.LIST, 2)
		if err != nil {
			return err
		}
		err = _os.Write_int32(int32(len(Vf)), 0)
		if err != nil {
			return err
		}
		for _, v := range Vf {

			err = _os.Write_string(v, 0)
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("func mismatch")
	}
	var status map[string]string
	*resp = requestf.ResponsePacket{
		IVersion:     1,
		CPacketType:  0,
		IRequestId:   req.IRequestId,
		IMessageType: 0,
		IRet:         0,
		SBuffer:      tools.ByteToInt8(_os.ToBytes()),
		Status:       status,
		SResultDesc:  "",
		Context:      req.Context,
	}
	_ = length
	_ = have
	_ = ty
	return nil
}
