package main

import (
	"fmt"
	"io/ioutil"
	"path"
	"sort"
	"strconv"
	"strings"
)

// VarType contains variable type(token)
type VarType struct {
	Type     TK       // basic type
	Unsigned bool     // whether unsigned
	TypeSt   string   // custom type name, such as an enumerated struct,at this time Type=tkName
	CType    TK       // make sure which type of custom type is,tkEnum, tkStruct
	TypeK    *VarType // vector's member variable,the key of map
	TypeV    *VarType // the value of map
	TypeL    int64    // lenth of array
}

// StructMember member struct.
type StructMember struct {
	Tag       int32
	Require   bool
	Type      *VarType
	Key       string // after the uppercase converted key
	OriginKey string // original key
	Default   string
	DefType   TK
}

// StructMemberSorter When serializing, make sure the tags are ordered.
type StructMemberSorter []StructMember

func (a StructMemberSorter) Len() int           { return len(a) }
func (a StructMemberSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StructMemberSorter) Less(i, j int) bool { return a[i].Tag < a[j].Tag }

//StructInfo record struct information.
type StructInfo struct {
	Name                string
	OriginName          string //original name
	Mb                  []StructMember
	DependModule        map[string]bool
	DependModuleWithJce map[string]string
}

//ArgInfo record argument information.
type ArgInfo struct {
	Name       string
	OriginName string //original name
	IsOut      bool
	Type       *VarType
}

//FunInfo record function information.
type FunInfo struct {
	Name       string // after the uppercase converted name
	OriginName string // original name
	HasRet     bool
	RetType    *VarType
	Args       []ArgInfo
}

//InterfaceInfo record interface information.
type InterfaceInfo struct {
	Name                string
	OriginName          string // original name
	Fun                 []FunInfo
	DependModule        map[string]bool
	DependModuleWithJce map[string]string
}

//EnumMember record member information.
type EnumMember struct {
	Key   string
	Type  int
	Value int32  //type 0
	Name  string //type 1
}

//EnumInfo record EnumMember information include name.
type EnumInfo struct {
	Module     string
	Name       string
	OriginName string // original name
	Mb         []EnumMember
}

//ConstInfo record const information.
type ConstInfo struct {
	Type       *VarType
	Name       string
	OriginName string // original name
	Value      string
}

//HashKeyInfo record hashkey information.
type HashKeyInfo struct {
	Name   string
	Member []string
}

//Parse record information of parse file.
type Parse struct {
	Source string

	Module       string
	OriginModule string
	Include      []string

	Struct    []StructInfo
	Interface []InterfaceInfo
	Enum      []EnumInfo
	Const     []ConstInfo
	HashKey   []HashKeyInfo

	// have parsed include file
	IncParse []*Parse

	lex   *LexState
	t     *Token
	lastT *Token

	// jce include chain
	IncChain []string

	// proto file name(not include .tars)
	ProtoName string

	DependModuleWithJce map[string]bool
}

func (p *Parse) parseErr(err string) {
	line := "0"
	if p.t != nil {
		line = strconv.Itoa(p.t.Line)
	}

	panic(p.Source + ": " + line + ". " + err)
}

func (p *Parse) next() {
	p.lastT = p.t
	p.t = p.lex.NextToken()
}

func (p *Parse) expect(t TK) {
	p.next()
	if p.t.T != t {
		p.parseErr("expect " + TokenMap[t])
	}
}

func (p *Parse) makeUnsigned(utype *VarType) {
	switch utype.Type {
	case tkTInt, tkTShort, tkTByte:
		utype.Unsigned = true
	default:
		p.parseErr("type " + TokenMap[utype.Type] + " unsigned decoration is not supported")
	}
}

func (p *Parse) parseType() *VarType {
	vtype := &VarType{Type: p.t.T}

	switch vtype.Type {
	case tkName:
		vtype.TypeSt = p.t.S.S
	case tkTInt, tkTBool, tkTShort, tkTLong, tkTByte, tkTFloat, tkTDouble, tkTString:
		// no nothing
	case tkTVector:
		p.expect(tkShl)
		p.next()
		vtype.TypeK = p.parseType()
		p.expect(tkShr)
	case tkTMap:
		p.expect(tkShl)
		p.next()
		vtype.TypeK = p.parseType()
		p.expect(tkComma)
		p.next()
		vtype.TypeV = p.parseType()
		p.expect(tkShr)
	case tkUnsigned:
		p.next()
		utype := p.parseType()
		p.makeUnsigned(utype)
		return utype
	default:
		p.parseErr("expert type")
	}
	return vtype
}

func (p *Parse) parseEnum() {
	enum := EnumInfo{}
	p.expect(tkName)
	enum.Name = p.t.S.S
	for _, v := range p.Enum {
		if v.Name == enum.Name {
			p.parseErr(enum.Name + " Redefine.")
		}
	}
	p.expect(tkBracel)

LFOR:
	for {
		p.next()
		switch p.t.T {
		case tkBracer:
			break LFOR
		case tkName:
			k := p.t.S.S
			p.next()
			switch p.t.T {
			case tkComma:
				m := EnumMember{Key: k, Type: 2}
				enum.Mb = append(enum.Mb, m)
			case tkBracer:
				m := EnumMember{Key: k, Type: 2}
				enum.Mb = append(enum.Mb, m)
				break LFOR
			case tkEq:
				p.next()
				switch p.t.T {
				case tkInteger:
					m := EnumMember{Key: k, Value: int32(p.t.S.I)}
					enum.Mb = append(enum.Mb, m)
				case tkName:
					m := EnumMember{Key: k, Type: 1, Name: p.t.S.S}
					enum.Mb = append(enum.Mb, m)
				default:
					p.parseErr("not expect " + TokenMap[p.t.T])
				}
				p.next()
				if p.t.T == tkBracer {
					break LFOR
				} else if p.t.T == tkComma {
				} else {
					p.parseErr("expect , or }")
				}
			}
		}
	}
	p.expect(tkSemi)
	p.Enum = append(p.Enum, enum)
}

func (p *Parse) parseStructMemberDefault(m *StructMember) {
	m.DefType = p.t.T
	switch p.t.T {
	case tkInteger:
		if !isNumberType(m.Type.Type) && m.Type.Type != tkName {
			// enum auto defined type ,default value is number.
			p.parseErr("type does not accept number")
		}
		m.Default = p.t.S.S
	case tkFloat:
		if !isNumberType(m.Type.Type) {
			p.parseErr("type does not accept number")
		}
		m.Default = p.t.S.S
	case tkString:
		if isNumberType(m.Type.Type) {
			p.parseErr("type does not accept string")
		}
		m.Default = `"` + p.t.S.S + `"`
	case tkTrue:
		if m.Type.Type != tkTBool {
			p.parseErr("default value format error")
		}
		m.Default = "true"
	case tkFalse:
		if m.Type.Type != tkTBool {
			p.parseErr("default value format error")
		}
		m.Default = "false"
	case tkName:
		m.Default = p.t.S.S
	default:
		p.parseErr("default value format error")
	}
}

func (p *Parse) parseStructMember() *StructMember {
	// tag or end
	p.next()
	if p.t.T == tkBracer {
		return nil
	}
	if p.t.T != tkInteger {
		p.parseErr("expect tags.")
	}
	m := &StructMember{}
	m.Tag = int32(p.t.S.I)

	// require or optional
	p.next()
	if p.t.T == tkRequire {
		m.Require = true
	} else if p.t.T == tkOptional {
		m.Require = false
	} else {
		p.parseErr("expect require or optional")
	}

	// type
	p.next()
	if !isType(p.t.T) && p.t.T != tkName && p.t.T != tkUnsigned {
		p.parseErr("expect type")
	} else {
		m.Type = p.parseType()
	}

	// key
	p.expect(tkName)
	m.Key = p.t.S.S

	p.next()
	if p.t.T == tkSemi {
		return m
	}
	if p.t.T == tkSquarel {
		p.expect(tkInteger)
		m.Type = &VarType{Type: tkTArray, TypeK: m.Type, TypeL: p.t.S.I}
		p.expect(tkSquarer)
		p.expect(tkSemi)
		return m
	}
	if p.t.T != tkEq {
		p.parseErr("expect ; or =")
	}
	if p.t.T == tkTMap || p.t.T == tkTVector || p.t.T == tkName {
		p.parseErr("map, vector, custom type cannot set default value")
	}

	// default
	p.next()
	p.parseStructMemberDefault(m)
	p.expect(tkSemi)

	return m
}

func (p *Parse) checkTag(st *StructInfo) {
	set := make(map[int32]bool)
	for _, v := range st.Mb {
		if set[v.Tag] {
			p.parseErr("tag = " + strconv.Itoa(int(v.Tag)) + ". have duplicates")
		}
		set[v.Tag] = true
	}
}

func (p *Parse) sortTag(st *StructInfo) {
	sort.Sort(StructMemberSorter(st.Mb))
}

func (p *Parse) parseStruct() {
	st := StructInfo{}
	p.expect(tkName)
	st.Name = p.t.S.S
	for _, v := range p.Struct {
		if v.Name == st.Name {
			p.parseErr(st.Name + " Redefine.")
		}
	}
	p.expect(tkBracel)

	for {
		m := p.parseStructMember()
		if m == nil {
			break
		}
		st.Mb = append(st.Mb, *m)
	}
	p.expect(tkSemi) //semicolon at the end of the struct.

	p.checkTag(&st)
	p.sortTag(&st)

	p.Struct = append(p.Struct, st)
}

func (p *Parse) parseInterfaceFun() *FunInfo {
	fun := &FunInfo{}
	p.next()
	if p.t.T == tkBracer {
		return nil
	}
	if p.t.T == tkVoid {
		fun.HasRet = false
	} else if !isType(p.t.T) && p.t.T != tkName && p.t.T != tkUnsigned {
		p.parseErr("expect type")
	} else {
		fun.HasRet = true
		fun.RetType = p.parseType()
	}
	p.expect(tkName)
	fun.Name = p.t.S.S
	p.expect(tkPtl)

	p.next()
	if p.t.T == tkShr {
		return fun
	}

	// No parameter function, exit directly.
	if p.t.T == tkPtr {
		p.expect(tkSemi)
		return fun
	}

	for {
		arg := &ArgInfo{}
		if p.t.T == tkOut {
			arg.IsOut = true
			p.next()
		} else {
			arg.IsOut = false
		}

		arg.Type = p.parseType()
		p.next()
		if p.t.T == tkName {
			arg.Name = p.t.S.S
			p.next()
		}

		fun.Args = append(fun.Args, *arg)

		if p.t.T == tkComma {
			p.next()
		} else if p.t.T == tkPtr {
			p.expect(tkSemi)
			break
		} else {
			p.parseErr("expect , or )")
		}
	}
	return fun
}

func (p *Parse) parseInterface() {
	itf := &InterfaceInfo{}
	p.expect(tkName)
	itf.Name = p.t.S.S
	for _, v := range p.Interface {
		if v.Name == itf.Name {
			p.parseErr(itf.Name + " Redefine.")
		}
	}
	p.expect(tkBracel)

	for {
		fun := p.parseInterfaceFun()
		if fun == nil {
			break
		}
		itf.Fun = append(itf.Fun, *fun)
	}
	p.expect(tkSemi) //semicolon at the end of struct.
	p.Interface = append(p.Interface, *itf)
}

func (p *Parse) parseConst() {
	m := ConstInfo{}

	// type
	p.next()
	switch p.t.T {
	case tkTVector, tkTMap:
		p.parseErr("const no supports type vector or map.")
	case tkTBool, tkTByte, tkTShort,
		tkTInt, tkTLong, tkTFloat,
		tkTDouble, tkTString, tkUnsigned:
		m.Type = p.parseType()
	default:
		p.parseErr("expect type.")
	}

	p.expect(tkName)
	m.Name = p.t.S.S

	p.expect(tkEq)

	// default
	p.next()
	switch p.t.T {
	case tkInteger, tkFloat:
		if !isNumberType(m.Type.Type) {
			p.parseErr("type does not accept number")
		}
		m.Value = p.t.S.S
	case tkString:
		if isNumberType(m.Type.Type) {
			p.parseErr("type does not accept string")
		}
		m.Value = `"` + p.t.S.S + `"`
	case tkTrue:
		if m.Type.Type != tkTBool {
			p.parseErr("default value format error")
		}
		m.Value = "true"
	case tkFalse:
		if m.Type.Type != tkTBool {
			p.parseErr("default value format error")
		}
		m.Value = "false"
	default:
		p.parseErr("default value format error")
	}
	p.expect(tkSemi)

	p.Const = append(p.Const, m)
}

func (p *Parse) parseHashKey() {
	hashKey := HashKeyInfo{}
	p.expect(tkSquarel)
	p.expect(tkName)
	hashKey.Name = p.t.S.S
	p.expect(tkComma)
	for {
		p.expect(tkName)
		hashKey.Member = append(hashKey.Member, p.t.S.S)
		p.next()
		t := p.t
		switch t.T {
		case tkSquarer:
			p.expect(tkSemi)
			p.HashKey = append(p.HashKey, hashKey)
			return
		case tkComma:
		default:
			p.parseErr("expect ] or ,")
		}
	}
}

func (p *Parse) parseModuleSegment() {
	p.expect(tkBracel)

	for {
		p.next()
		t := p.t
		switch t.T {
		case tkBracer:
			p.expect(tkSemi)
			return
		case tkConst:
			p.parseConst()
		case tkEnum:
			p.parseEnum()
		case tkStruct:
			p.parseStruct()
		case tkInterface:
			p.parseInterface()
		case tkKey:
			p.parseHashKey()
		default:
			p.parseErr("not except " + TokenMap[t.T])
		}
	}
}

func (p *Parse) parseModule() {
	p.expect(tkName)

	if p.Module != "" {
		p.parseErr("do not repeat define module")
	}
	p.Module = p.t.S.S

	p.parseModuleSegment()
}

func (p *Parse) parseInclude() {
	p.expect(tkString)
	p.Include = append(p.Include, p.t.S.S)
}

// Looking for the true type of user-defined identifier
func (p *Parse) findTNameType(tname string) (TK, string, string) {
	for _, v := range p.Struct {
		if p.Module+"::"+v.Name == tname {
			return tkStruct, p.Module, p.ProtoName
		}
	}

	for _, v := range p.Enum {
		if p.Module+"::"+v.Name == tname {
			return tkEnum, p.Module, p.ProtoName
		}
	}

	for _, pInc := range p.IncParse {
		ret, mod, protoName := pInc.findTNameType(tname)
		if ret != tkName {
			return ret, mod, protoName
		}
	}
	// not find
	return tkName, p.Module, p.ProtoName
}

func (p *Parse) findEnumName(ename string) (*EnumMember, *EnumInfo) {
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
				p.parseErr(ename + " name conflict [" + cenum.Name + "::" + cmb.Key + " or " + enum.Name + "::" + mb.Key)
				return nil, nil
			}
		}
	}
	for _, pInc := range p.IncParse {
		if cmb == nil {
			cmb, cenum = pInc.findEnumName(ename)
		} else {
			break
		}
	}
	if cenum != nil && cenum.Module == "" {
		if *gModuleCycle == true {
			cenum.Module = p.ProtoName + "_" + p.Module
		} else {
			cenum.Module = p.Module
		}
	}
	return cmb, cenum
}

func addToSet(m *map[string]bool, module string) {
	if *m == nil {
		*m = make(map[string]bool)
	}
	(*m)[module] = true
}

func addToMap(m *map[string]string, module string, value string) {
	if *m == nil {
		*m = make(map[string]string)
	}
	(*m)[module] = value
}

func (p *Parse) checkDepTName(ty *VarType, dm *map[string]bool, dmj *map[string]string) {
	if ty.Type == tkName {
		name := ty.TypeSt
		if strings.Count(name, "::") == 0 {
			name = p.Module + "::" + name
		}

		mod := ""
		protoName := ""
		ty.CType, mod, protoName = p.findTNameType(name)
		if ty.CType == tkName {
			p.parseErr(ty.TypeSt + " not find define")
		}
		if *gModuleCycle == true {
			if mod != p.Module || protoName != p.ProtoName {
				var modStr string
				if *gModuleUpper {
					modStr = upperFirstLetter(mod)
				} else {
					modStr = mod
				}
				addToMap(dmj, protoName+"_"+modStr, protoName)

				if strings.Contains(ty.TypeSt, mod+"::") {
					ty.TypeSt = strings.Replace(ty.TypeSt, mod+"::", protoName+"_"+modStr+"::", 1)
				} else {
					ty.TypeSt = protoName + "_" + modStr + "::" + ty.TypeSt
				}
			} else {
				// the same Module ,do not add self.
				ty.TypeSt = strings.Replace(ty.TypeSt, mod+"::", "", 1)
			}
		} else {
			if mod != p.Module {
				addToSet(dm, mod)
			} else {
				// the same Module ,do not add self.
				ty.TypeSt = strings.Replace(ty.TypeSt, mod+"::", "", 1)
			}
		}
	} else if ty.Type == tkTVector {
		p.checkDepTName(ty.TypeK, dm, dmj)
	} else if ty.Type == tkTMap {
		p.checkDepTName(ty.TypeK, dm, dmj)
		p.checkDepTName(ty.TypeV, dm, dmj)
	}
}

// analysis custom type，whether have definition
func (p *Parse) analyzeTName() {
	for i, v := range p.Struct {
		for _, v := range v.Mb {
			ty := v.Type
			p.checkDepTName(ty, &p.Struct[i].DependModule, &p.Struct[i].DependModuleWithJce)
		}
	}

	for i, v := range p.Interface {
		for _, v := range v.Fun {
			for _, v := range v.Args {
				ty := v.Type
				p.checkDepTName(ty, &p.Interface[i].DependModule, &p.Interface[i].DependModuleWithJce)
			}
			if v.RetType != nil {
				p.checkDepTName(v.RetType, &p.Interface[i].DependModule, &p.Interface[i].DependModuleWithJce)
			}
		}
	}
}

func (p *Parse) analyzeDefault() {
	for _, v := range p.Struct {
		for i, r := range v.Mb {
			if r.Default != "" && r.DefType == tkName {
				mb, enum := p.findEnumName(r.Default)
				if mb == nil || enum == nil {
					p.parseErr("can not find default value" + r.Default)
				}
				defValue := enum.Name + "_" + upperFirstLetter(mb.Key)
				var currModule string
				if *gModuleCycle == true {
					currModule = p.ProtoName + "_" + p.Module
				} else {
					currModule = p.Module
				}
				if len(enum.Module) > 0 && currModule != enum.Module {
					defValue = enum.Module + "." + defValue
				}
				v.Mb[i].Default = defValue
			}
		}
	}
}

// TODO analysis key[]，have quoted the correct struct and member name.
func (p *Parse) analyzeHashKey() {

}

func (p *Parse) analyzeDepend() {
	for _, v := range p.Include {
		//#include support relative path,example: ../test.tars
		relativePath := path.Dir(p.Source)
		dependFile := relativePath + "/" + v
		pInc := ParseFile(dependFile, p.IncChain)
		p.IncParse = append(p.IncParse, pInc)
		fmt.Println("parse include: ", v)
	}

	p.analyzeDefault()
	p.analyzeTName()
	p.analyzeHashKey()
}

func (p *Parse) parse() {
OUT:
	for {
		p.next()
		t := p.t
		switch t.T {
		case tkEos:
			break OUT
		case tkInclude:
			p.parseInclude()
		case tkModule:
			p.parseModule()
		default:
			p.parseErr("Expect include or module.")
		}
	}
	p.analyzeDepend()
}

func newParse(s string, b []byte, incChain []string) *Parse {
	p := &Parse{Source: s, ProtoName: path2ProtoName(s)}
	for _, v := range incChain {
		if s == v {
			panic("jce circular reference: " + s)
		}
	}
	incChain = append(incChain, s)
	p.IncChain = incChain
	fmt.Println(s, p.IncChain)

	p.lex = NewLexState(s, b)
	return p
}

//ParseFile parse a file,return grammer tree.
func ParseFile(path string, incChain []string) *Parse {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("file read error: " + path + ". " + err.Error())
	}

	p := newParse(path, b, incChain)
	p.parse()

	return p
}
