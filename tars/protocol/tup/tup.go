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
	data map[string][]byte
	//os 		codec.Buffer
	//is 		codec.Reader
}

func NewUniAttribute() *UniAttribute {
	return &UniAttribute{data: make(map[string][]byte)}
}

func (u *UniAttribute) PutBuffer(k string, buf []byte) {
	u.data[k] = make([]byte, len(buf))
	copy(u.data[k], buf)
}

func (u *UniAttribute) GetBuffer(k string, buf *[]byte) error {
	var (
		err error
		ok  bool
	)
	if *buf, ok = u.data[k]; !ok {
		err = fmt.Errorf("tup get error: donot find key: %s", k)
	}

	return err
}

func (u *UniAttribute) Encode(os *codec.Buffer) error {
	err := os.WriteHead(codec.MAP, 0)
	if err != nil {
		return err
	}
	err = os.WriteInt32(int32(len(u.data)), 0)
	if err != nil {
		return err
	}
	for k, v := range u.data {
		err = os.WriteString(k, 0)
		if err != nil {
			return err
		}

		err = os.WriteHead(codec.SimpleList, 1)
		if err != nil {
			return err
		}
		err = os.WriteHead(codec.BYTE, 0)
		if err != nil {
			return err
		}
		err = os.WriteInt32(int32(len(v)), 0)
		if err != nil {
			return err
		}
		err = os.WriteBytes(v)
		if err != nil {
			return err
		}
	}

	return err
}

func (u *UniAttribute) Decode(is *codec.Reader) error {
	var (
		have bool
		ty   byte
		err  error
	)
	_, err = is.SkipTo(codec.MAP, 0, false)
	if err != nil {
		return err
	}

	var length int32 = 0
	err = is.ReadInt32(&length, 0, true)
	if err != nil {
		return err
	}

	for i, e := int32(0), length; i < e; i++ {
		var k string
		var v []byte

		err = is.ReadString(&k, 0, false)
		if err != nil {
			return err
		}

		have, ty, err = is.SkipToNoCheck(1, false)
		if err != nil {
			return err
		}
		if have {
			if ty == codec.SimpleList {
				_, err = is.SkipTo(codec.BYTE, 0, true)
				if err != nil {
					return err
				}

				var byteLen int32 = 0
				err = is.ReadInt32(&byteLen, 0, true)
				if err != nil {
					return err
				}

				err = is.ReadBytes(&v, byteLen, true)
				if err != nil {
					return err
				}

				u.data[k] = v
			} else {
				err = fmt.Errorf("require vector, but not")
				if err != nil {
					return err
				}
			}
		}
	}

	return err
}

func (u *UniAttribute) putBase(data interface{}, os *codec.Buffer) error {
	var err error
	//os := codec.NewBuffer()
	switch d := data.(type) {
	case int64:
		err = os.WriteInt64(d, 0)
	case int32:
		err = os.WriteInt32(d, 0)
	case int16:
		err = os.WriteInt16(d, 0)
	case int8:
		err = os.WriteInt8(d, 0)
	case uint32:
		err = os.WriteUint32(d, 0)
	case uint16:
		err = os.WriteUint16(d, 0)
	case uint8:
		err = os.WriteUint8(d, 0)
	case bool:
		err = os.WriteBool(d, 0)
	case float64:
		err = os.WriteFloat64(d, 0)
	case float32:
		err = os.WriteFloat32(d, 0)
	case string:
		err = os.WriteString(d, 0)
	case TarsStructIF:
		err = data.(TarsStructIF).WriteBlock(os, 0)
	default:
		err = fmt.Errorf("tup put error: not support type")
	}

	return err
}

func (u *UniAttribute) doPut(data interface{}, os *codec.Buffer) error {
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
			err = os.WriteInt32(int32(0), 0)
			// if err != nil {
			// 	return err
			// }
			// err = fmt.Errorf("Error Array Len:0")
			return err
		}

		switch s.Index(0).Interface().(type) {
		case int8:
			err = os.WriteHead(codec.SimpleList, 0)
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.BYTE, 0)
			if err != nil {
				return err
			}
			err = os.WriteInt32(int32(s.Len()), 0)
			if err != nil {
				return err
			}
			err = os.WriteSliceInt8(data.([]int8))
			if err != nil {
				return err
			}
		case uint8:
			err = os.WriteHead(codec.SimpleList, 0)
			if err != nil {
				return err
			}
			err = os.WriteHead(codec.BYTE, 0)
			if err != nil {
				return err
			}
			err = os.WriteInt32(int32(s.Len()), 0)
			if err != nil {
				return err
			}
			err = os.WriteSliceUint8(data.([]uint8))
			if err != nil {
				return err
			}
		default:
			err = os.WriteHead(codec.LIST, 0)
			if err != nil {
				return err
			}
			err = os.WriteInt32(int32(s.Len()), 0)
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

func (u *UniAttribute) Put(k string, data interface{}) error {
	var err error
	os := codec.NewBuffer()
	err = u.doPut(data, os)

	if err == nil {
		u.data[k] = os.ToBytes()
		fmt.Printf("%s = %d \n", k, len(os.ToBytes()))
	}
	return err
}

func (u *UniAttribute) getBase(data interface{}, is *codec.Reader) error {
	var err error
	// if v, ok := u.data[k]; ok {
	// 	is := codec.NewReader(v)
	switch d := data.(type) {
	case *int64:
		err = is.ReadInt64(d, 0, true)
	case *int32:
		err = is.ReadInt32(d, 0, true)
	case *int16:
		err = is.ReadInt16(d, 0, true)
	case *int8:
		err = is.ReadInt8(d, 0, true)
	case *uint32:
		err = is.ReadUint32(d, 0, true)
	case *uint16:
		err = is.ReadUint16(d, 0, true)
	case *uint8:
		err = is.ReadUint8(d, 0, true)
	case *bool:
		err = is.ReadBool(d, 0, true)
	case *float64:
		err = is.ReadFloat64(d, 0, true)
	case *float32:
		err = is.ReadFloat32(d, 0, true)
	case *string:
		err = is.ReadString(d, 0, true)
	case TarsStructIF:
		err = data.(TarsStructIF).ReadBlock(is, 0, true)
	default:
		err = fmt.Errorf("tup get error: not support type")
	}
	// } else {
	// 	err = fmt.Errorf("Tup Get Error: donot find key: %s!", k)
	// }

	return err
}

func (u *UniAttribute) DoGet(data interface{}, is *codec.Reader) error {
	var err error
	// vOF := reflect.ValueOf(data).Elem()
	switch reflect.TypeOf(data).Kind() {
	case reflect.Slice:
		fmt.Println("get slice ...")

		have, ty, err := is.SkipToNoCheck(0, false)
		if err != nil {
			return err
		}
		if have {
			if ty == codec.LIST {
				var length int32
				err = is.ReadInt32(&length, 0, true)
				if err != nil {
					return err
				}

				// st.Vf = make([]float32, length, length)
				// for i0, e0 := int32(0), length; i0 < e0; i0++ {
				//
				// 	 err = _is.Read_float32(&st.Vf[i0], 0, false)
				//	 if err != nil {
				//		return err
				//	 }
				// }
			} else if ty == codec.SimpleList {
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
	if v, ok := u.data[k]; ok {
		//is := codec.NewReader(v)
		//err = u.doGet(data, is)
		err = fmt.Errorf("tup not support! Please use GetBuffer()")
		_ = v
	} else {
		err = fmt.Errorf("tup get error: donot find key: %s", k)
	}

	return err
}
