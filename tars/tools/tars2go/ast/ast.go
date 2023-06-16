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

// StructInfo record struct information.
type StructInfo struct {
	Name                string
	OriginName          string //original name
	Mb                  []StructMember
	DependModule        map[string]bool
	DependModuleWithJce map[string]string
}

// ArgInfo record argument information.
type ArgInfo struct {
	Name       string
	OriginName string //original name
	IsOut      bool
	Type       *VarType
}

// FunInfo record function information.
type FunInfo struct {
	Name       string // after the uppercase converted name
	OriginName string // original name
	HasRet     bool
	RetType    *VarType
	Args       []ArgInfo
}

// InterfaceInfo record interface information.
type InterfaceInfo struct {
	Name                string
	OriginName          string // original name
	Fun                 []FunInfo
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

// EnumInfo record EnumMember information include name.
type EnumInfo struct {
	Module     string
	Name       string
	OriginName string // original name
	Mb         []EnumMember
}

// ConstInfo record const information.
type ConstInfo struct {
	Type       *VarType
	Name       string
	OriginName string // original name
	Value      string
}

// HashKeyInfo record hash key information.
type HashKeyInfo struct {
	Name   string
	Member []string
}

type ModuleInfo struct {
	Source string
	// proto file name(not include .tars)
	ProtoName  string
	Name       string
	OriginName string
	Include    []string

	Struct    []StructInfo
	HashKey   []HashKeyInfo
	Enum      []EnumInfo
	Const     []ConstInfo
	Interface []InterfaceInfo

	// have parsed include file
	IncModule []*ModuleInfo
}

// Rename module
func (p *ModuleInfo) Rename(moduleUpper bool) {
	p.OriginName = p.Name
	if moduleUpper {
		p.Name = utils.UpperFirstLetter(p.Name)
	}
}

// FindTNameType Looking for the true type of user-defined identifier
func (p *ModuleInfo) FindTNameType(tname string) (token.Type, string, string) {
	for _, v := range p.Struct {
		if p.Name+"::"+v.Name == tname {
			return token.Struct, p.Name, p.ProtoName
		}
	}

	for _, v := range p.Enum {
		if p.Name+"::"+v.Name == tname {
			return token.Enum, p.Name, p.ProtoName
		}
	}

	for _, pInc := range p.IncModule {
		ret, mod, protoName := pInc.FindTNameType(tname)
		if ret != token.Name {
			return ret, mod, protoName
		}
	}
	// not find
	return token.Name, p.Name, p.ProtoName
}

func (p *ModuleInfo) FindEnumName(ename string, moduleCycle bool) (*EnumMember, *EnumInfo, error) {
	if strings.Contains(ename, "::") {
		vec := strings.Split(ename, "::")
		if len(vec) >= 2 {
			ename = vec[1]
		}
	}
	var cmb *EnumMember
	var cenum *EnumInfo
	for ek, enum := range p.Enum {
		for mk, mb := range enum.Mb {
			if mb.Key != ename {
				continue
			}
			if cmb == nil {
				cmb = &enum.Mb[mk]
				cenum = &p.Enum[ek]
			} else {
				return nil, nil, errors.New(ename + " name conflict [" + cenum.Name + "::" + cmb.Key + " or " + enum.Name + "::" + mb.Key)
			}
		}
	}
	var err error
	for _, pInc := range p.IncModule {
		if cmb == nil {
			cmb, cenum, err = pInc.FindEnumName(ename, moduleCycle)
			if err != nil {
				return cmb, cenum, err
			}
		} else {
			break
		}
	}
	if cenum != nil && cenum.Module == "" {
		if moduleCycle {
			cenum.Module = p.ProtoName + "_" + p.Name
		} else {
			cenum.Module = p.Name
		}
	}
	return cmb, cenum, nil
}

// Rename struct
// struct Name { 1 require Mb type}
func (st *StructInfo) Rename() {
	st.OriginName = st.Name
	st.Name = utils.UpperFirstLetter(st.Name)
	for i := range st.Mb {
		st.Mb[i].OriginKey = st.Mb[i].Key
		st.Mb[i].Key = utils.UpperFirstLetter(st.Mb[i].Key)
	}
}

// Rename interface
// interface Name { Fun }
func (itf *InterfaceInfo) Rename() {
	itf.OriginName = itf.Name
	itf.Name = utils.UpperFirstLetter(itf.Name)
	for i := range itf.Fun {
		itf.Fun[i].Rename()
	}
}

func (en *EnumInfo) Rename() {
	en.OriginName = en.Name
	en.Name = utils.UpperFirstLetter(en.Name)
	for i := range en.Mb {
		en.Mb[i].Key = utils.UpperFirstLetter(en.Mb[i].Key)
	}
}

func (cst *ConstInfo) Rename() {
	cst.OriginName = cst.Name
	cst.Name = utils.UpperFirstLetter(cst.Name)
}

// Rename func
// type Fun (arg ArgType), in case keyword and name conflicts,argname need to capitalize.
// Fun (type int32)
func (fun *FunInfo) Rename() {
	fun.OriginName = fun.Name
	fun.Name = utils.UpperFirstLetter(fun.Name)
	for i := range fun.Args {
		fun.Args[i].OriginName = fun.Args[i].Name
		// func args donot upper firs
		//fun.Args[i].Name = utils.UpperFirstLetter(fun.Args[i].Name)
	}
}
