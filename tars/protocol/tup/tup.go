package tup

import (
	"fmt"
	"reflect"

	"github.com/TarsCloud/TarsGo/tars/protocol/codec"
)

type TarsStructIF interface {
	WriteBlock(os *codec.Buffer, tag byte) error
    ReadBlock(is *codec.Reader, tag byte, require bool) error	
}

type UniAttribute struct {
	_data 	map[string][]byte
	//os 		codec.Buffer
	//is 		codec.Reader
}

func NewUniAttribute() *UniAttribute {
	return &UniAttribute{_data: make(map[string][]byte)}
}


func (u *UniAttribute) PutBuffer(k string, buf []byte) {
	u._data[k] = make([]byte, len(buf))
	copy(u._data[k], buf)
}

func (u *UniAttribute) GetBuffer(k string, buf *[]byte) error  {
	var err error
	var ok bool = false
	if *buf, ok = u._data[k]; !ok {
		err = fmt.Errorf("Tup Get Error: donot find key: %s!", k)
	}

	return err
}

func (u *UniAttribute) Encode(os *codec.Buffer) error {
	err := os.WriteHead(codec.MAP, 0)
	if err != nil {
		return err
	}
	err = os.Write_int32(int32(len(u._data)), 0)
	if err != nil {
		return err
	}
	for k, v := range u._data {
		err = os.Write_string(k, 0)
		if err != nil {
			return err
		}

		err = os.WriteHead(codec.SIMPLE_LIST, 1)
		if err != nil {
			return err
		}
		err = os.WriteHead(codec.BYTE, 0)
		if err != nil {
			return err
		}
		err = os.Write_int32(int32(len(v)), 0)
		if err != nil {
			return err
		}
		err = os.Write_bytes(v)
		if err != nil {
			return err
		}
	}

	return  err
}

func (u *UniAttribute) Decode(is *codec.Reader) error  {
	err, have := is.SkipTo(codec.MAP, 0, false)
	if err != nil {
		return err
	}

	var length int32 = 0
	err = is.Read_int32(&length, 0, true)
	if err != nil {
		return err
	}

	var ty byte
	for i, e := int32(0), length; i < e; i++ {
		var k string
		var v []byte

		err = is.Read_string(&k, 0, false)
		if err != nil {
			return err
		}

		err, have, ty = is.SkipToNoCheck(1, false)
		if err != nil {
			return err
		}
		if have {
			if ty == codec.SIMPLE_LIST {

				err, _ = is.SkipTo(codec.BYTE, 0, true)
				if err != nil {
					return err
				}
				var byteLen int32 = 0
				err = is.Read_int32(&byteLen, 0, true)
				if err != nil {
					return err
				}
				err = is.Read_bytes(&v, byteLen, true)
				if err != nil {
					return err
				}

				u._data[k] = v

			} else {
				err = fmt.Errorf("require vector, but not")
				if err != nil {
					return err
				}
			}
		}
	}

	return  err
}

func (u *UniAttribute) putBase(data interface{}, os *codec.Buffer) error  {
	var err error
	//os := codec.NewBuffer()
	switch data.(type) {
	case int64:
		err = os.Write_int64(data.(int64), 0)
	case int32:
		err = os.Write_int32(data.(int32), 0)
	case int16:
		err = os.Write_int16(data.(int16), 0)
	case int8:
		err = os.Write_int8(data.(int8), 0)
	case uint32:
		err = os.Write_uint32(data.(uint32), 0)
	case uint16:
		err = os.Write_uint16(data.(uint16), 0)
	case uint8:
		err = os.Write_uint8(data.(uint8), 0)
	case bool:
		err = os.Write_bool(data.(bool), 0)
	case float64:
		err = os.Write_float64(data.(float64), 0)
	case float32:
		err = os.Write_float32(data.(float32), 0)
	case string:
		err = os.Write_string(data.(string), 0)
	case TarsStructIF:
		err = data.(TarsStructIF).WriteBlock(os, 0)
	default:
		err = fmt.Errorf("Tup Put Error: not support type!")
	}
	
	return err
}

func (u *UniAttribute) doPut(data interface{}, os *codec.Buffer) error  {
	var err error
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice, reflect.Array:
		fmt.Println("Tup Put Array...")
		s := reflect.ValueOf(data)
		if s.Len() == 0 {
			err = os.WriteHead(codec.LIST, 0)
			if err != nil {
				return err
			}
			err = os.Write_int32(int32(0), 0)
			// if err != nil {
			// 	return err
			// }
			//err = fmt.Errorf("Error Array Len:0")
			return err
		}

		switch s.Index(0).Interface().(type) {
		case int8:
			err = os.WriteHead(codec.SIMPLE_LIST, 0)
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.BYTE, 0)
			if err != nil {
				return err
			}
			err = os.Write_int32(int32(s.Len()), 0)
			if err != nil {
				return err
			}
			err = os.Write_slice_int8(data.([]int8))
			if err != nil {
				return err
			}
		case uint8:
			err = os.WriteHead(codec.SIMPLE_LIST, 0)
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.BYTE, 0)
			if err != nil {
				return err
			}
			err = os.Write_int32(int32(s.Len()), 0)
			if err != nil {
				return err
			}
			err = os.Write_slice_uint8(data.([]uint8))
			if err != nil {
				return err
			}
		default:
			err = os.WriteHead(codec.LIST, 0)
			if err != nil {
				return err
			}
			err = os.Write_int32(int32(s.Len()), 0)
			if err != nil {
				return err
			}
			for i := 0; i < s.Len(); i++ {
				err = u.doPut(s.Index(i).Interface(), os)
				if err != nil {
					fmt.Println("error, data:", s.Index(i), err)
					break
				}
			}
		}
		
	default:
		err = u.putBase(data, os)
	}
	return err
}

func (u *UniAttribute) Put(k string, data interface{}) error  {
	var err error
	os := codec.NewBuffer()
	err = u.doPut(data, os)

	if err == nil {
		u._data[k] = os.ToBytes()
		fmt.Printf("%s = %d \n", k, len(os.ToBytes()))
	}
	return err
}

func (u *UniAttribute) getBase(data interface{}, is *codec.Reader) error {
	var err error
	// if v, ok := u._data[k]; ok {
	// 	is := codec.NewReader(v)
	switch (data).(type) {
	case *int64:
		err = is.Read_int64(data.(*int64), 0, true)
	case *int32:
		err = is.Read_int32(data.(*int32), 0, true)
	case *int16:
		err = is.Read_int16(data.(*int16), 0, true)
	case *int8:
		err = is.Read_int8(data.(*int8), 0, true)
	case *uint32:
		err = is.Read_uint32(data.(*uint32), 0, true)
	case *uint16:
		err = is.Read_uint16(data.(*uint16), 0, true)
	case *uint8:
		err = is.Read_uint8(data.(*uint8), 0, true)
	case *bool:
		err = is.Read_bool(data.(*bool), 0, true)
	case *float64:
		err = is.Read_float64(data.(*float64), 0, true)
	case *float32:
		err = is.Read_float32(data.(*float32), 0, true)
	case *string:
		err = is.Read_string((data).(*string), 0, true)
	case TarsStructIF:
		err = data.(TarsStructIF).ReadBlock(is, 0, true)
	default:
		err = fmt.Errorf("Tup get error: not support type!")
	}
	// } else {
	// 	err = fmt.Errorf("Tup Get Error: donot find key: %s!", k)
	// }
	
	return err
}

func (u *UniAttribute)doGet(data interface{}, is *codec.Reader) error {
	var err error
	//vOF := reflect.ValueOf(data).Elem()
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		fmt.Println("get slice ...")
		
		err, have, ty := is.SkipToNoCheck(0, false)
		if err != nil {
			return err
		}
		if have {
			if ty == codec.LIST {
				var length int32
				err = is.Read_int32(&length, 0, true)
				if err != nil {
					return err
				}

				//st.Vf = make([]float32, length, length)
				//for i0, e0 := int32(0), length; i0 < e0; i0++ {
				//
				//	err = _is.Read_float32(&st.Vf[i0], 0, false)
				//	if err != nil {
				//		return err
				//	}
				//}
			} else if ty == codec.SIMPLE_LIST {
				err = fmt.Errorf("not support simple_list type")
				if err != nil {
					return err
				}
			} else {
				err = fmt.Errorf("require vector, but not")
				if err != nil {
					return err
				}
			}
		}

	default:
		err = u.getBase(data, is)
	}
	return err
}

func (u *UniAttribute) Get(k string, data interface{}) error {
	var err error
	if v, ok := u._data[k]; ok {
		//is := codec.NewReader(v)
		//err = u.doGet(data, is)
		err = fmt.Errorf("Tup not support! Please use GetBuffer()")
		_ = v
	} else {
		err = fmt.Errorf("Tup Get Error: donot find key: %s!", k)
	}

	return err
}
