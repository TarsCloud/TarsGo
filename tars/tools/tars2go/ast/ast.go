package ast

import (
	"errors"
	"strings"

	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/token"
	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/utils"
)

// VarType contains variable type(token)
type VarType struct {
	Type     token.Type // basic type
	Unsigned bool       // whether unsigned
	TypeSt   string     // custom type name, such as an enumerated struct,at this time Type=token.Name
	CType    token.Type // make sure which type of custom type is,token.Enum, token.Struct
	TypeK    *VarType   // vector's member variable,the key of map
	TypeV    *VarType   // the value of map
	TypeL    int64      // length of array
}

// StructMember member struct.
type StructMember struct {
	Tag       int32
	Require   bool
	Type      *VarType
	Key       string // after the uppercase converted key
	OriginKey string // original key
	Default   string
	DefType   token.Type
}

// StructMemberSorter When serializing, make sure the tags are ordered.
type StructMemberSorter []StructMember

func (a StructMemberSorter) Len() int           { return len(a) }
func (a StructMemberSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StructMemberSorter) Less(i, j int) bool { return a[i].Tag < a[j].Tag }

// Struct record struct information.
type Struct struct {
	Name                string
	OriginName          string //original name
	Mb                  []StructMember
	DependModule        map[string]bool
	DependModuleWithJce map[string]string
}

// Arg record argument information.
type Arg struct {
	Name       string
	OriginName string //original name
	IsOut      bool
	Type       *VarType
}

// Func record function information.
type Func struct {
	Name       string // after the uppercase converted name
	OriginName string // original name
	HasRet     bool
	RetType    *VarType
	Args       []Arg
}

// Interface record interface information.
type Interface struct {
	Name                string
	OriginName          string // original name
	Funcs               []Func
	DependModule        map[string]bool
	DependModuleWithJce map[string]string
}

// EnumMember record member information.
type EnumMember struct {
	Key   string
	Type  int
	Value int32  //type 0
	Name  string //type 1
}

// Enum record EnumMember information include name.
type Enum struct {
	Module     string
	Name       string
	OriginName string // original name
	Mb         []EnumMember
}

// Const record const information.
type Const struct {
	Type       *VarType
	Name       string
	OriginName string // original name
	Value      string
}

// HashKey record hash key information.
type HashKey struct {
	Name   string
	Member []string
}

type TarsFile struct {
	Source string
	// proto file name(not include .tars)
	ProtoName string
	Module    Module

	Include []string
	// have parsed include file
	IncTarsFile []*TarsFile
}

type Module struct {
	Name       string
	OriginName string

	Struct    []Struct
	HashKey   []HashKey
	Enum      []Enum
	Const     []Const
	Interface []Interface
}

// Rename module
func (m *Module) Rename(moduleUpper bool) {
	m.OriginName = m.Name
	if moduleUpper {
		m.Name = utils.UpperFirstLetter(m.Name)
	}
}

// FindTNameType Looking for the true type of user-defined identifier
func (tf *TarsFile) FindTNameType(tName string) (token.Type, string, string) {
	for _, v := range tf.Module.Struct {
		if tf.Module.Name+"::"+v.Name == tName {
			return token.Struct, tf.Module.Name, tf.ProtoName
		}
	}

	for _, v := range tf.Module.Enum {
		if tf.Module.Name+"::"+v.Name == tName {
			return token.Enum, tf.Module.Name, tf.ProtoName
		}
	}

	for _, tarsFile := range tf.IncTarsFile {
		ret, mod, protoName := tarsFile.FindTNameType(tName)
		if ret != token.Name {
			return ret, mod, protoName
		}
	}
	// not find
	return token.Name, tf.Module.Name, tf.ProtoName
}

func (tf *TarsFile) FindEnumName(ename string, moduleCycle bool) (*EnumMember, *Enum, error) {
	if strings.Contains(ename, "::") {
		vec := strings.Split(ename, "::")
		if len(vec) >= 2 {
			ename = vec[1]
		}
	}
	var cmb *EnumMember
	var cenum *Enum
	for ek, enum := range tf.Module.Enum {
		for mk, mb := range enum.Mb {
			if mb.Key != ename {
				continue
			}
			if cmb == nil {
				cmb = &enum.Mb[mk]
				cenum = &tf.Module.Enum[ek]
			} else {
				return nil, nil, errors.New(ename + " name conflict [" + cenum.Name + "::" + cmb.Key + " or " + enum.Name + "::" + mb.Key)
			}
		}
	}
	var err error
	for _, tarsFile := range tf.IncTarsFile {
		if cmb == nil {
			cmb, cenum, err = tarsFile.FindEnumName(ename, moduleCycle)
			if err != nil {
				return cmb, cenum, err
			}
		} else {
			break
		}
	}
	if cenum != nil && cenum.Module == "" {
		if moduleCycle {
			cenum.Module = tf.ProtoName + "_" + tf.Module.Name
		} else {
			cenum.Module = tf.Module.Name
		}
	}
	return cmb, cenum, nil
}

// Rename Struct Name { 1 require Mb type}
func (st *Struct) Rename() {
	st.OriginName = st.Name
	st.Name = utils.UpperFirstLetter(st.Name)
	for i := range st.Mb {
		st.Mb[i].OriginKey = st.Mb[i].Key
		st.Mb[i].Key = utils.UpperFirstLetter(st.Mb[i].Key)
	}
}

// Rename Interface Name { Funcs }
func (itf *Interface) Rename() {
	itf.OriginName = itf.Name
	itf.Name = utils.UpperFirstLetter(itf.Name)
	for i := range itf.Funcs {
		itf.Funcs[i].Rename()
	}
}

// Rename Enum Name { Mb }
func (en *Enum) Rename() {
	en.OriginName = en.Name
	en.Name = utils.UpperFirstLetter(en.Name)
	for i := range en.Mb {
		en.Mb[i].Key = utils.UpperFirstLetter(en.Mb[i].Key)
	}
}

// Rename Const Name
func (cst *Const) Rename() {
	cst.OriginName = cst.Name
	cst.Name = utils.UpperFirstLetter(cst.Name)
}

// Rename Func Name { Args }
// type Funcs (arg ArgType), in case keyword and name conflicts, arg name need to capitalize.
// Funcs (type int32)
func (fun *Func) Rename() {
	fun.OriginName = fun.Name
	fun.Name = utils.UpperFirstLetter(fun.Name)
	for i := range fun.Args {
		fun.Args[i].OriginName = fun.Args[i].Name
		// func args do not upper firs
		// fun.Args[i].Name = utils.UpperFirstLetter(fun.Args[i].Name)
	}
}
