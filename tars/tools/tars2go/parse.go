package main

import (
	"fmt"
	"io/ioutil"
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
}

// StructMember member struct.
type StructMember struct {
	Tag     int32
	Require bool
	Type    *VarType
	Key     string // after the uppercase converted key
	KeyStr  string // original key
	Default string
	DefType TK
}

// StructMemberSorter When serializing, make sure the tags are ordered.
type StructMemberSorter []StructMember

func (a StructMemberSorter) Len() int           { return len(a) }
func (a StructMemberSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StructMemberSorter) Less(i, j int) bool { return a[i].Tag < a[j].Tag }

//StructInfo record struct information.
type StructInfo struct {
	TName        string
	Mb           []StructMember
	DependModule map[string]bool
}

//ArgInfo record argument information.
type ArgInfo struct {
	Name  string
	IsOut bool
	Type  *VarType
}

//FunInfo record function information.
type FunInfo struct {
	Name    string // after the uppercase converted name
	NameStr string // original name
	HasRet  bool
	RetType *VarType
	Args    []ArgInfo
}

//InterfaceInfo record interface information.
type InterfaceInfo struct {
	TName        string
	Fun          []FunInfo
	DependModule map[string]bool
}

//EnumMember record member information.
type EnumMember struct {
	Key   string
	Value int32
}

//EnumInfo record EnumMember information include name.
type EnumInfo struct {
	TName string
	Mb    []EnumMember
}

//ConstInfo record const information.
type ConstInfo struct {
	Type  *VarType
	Key   string
	Value string
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
	enum.TName = p.t.S.S
	for _, v := range p.Enum {
		if v.TName == enum.TName {
			p.parseErr(enum.TName + " Redefine.")
		}
	}
	p.expect(tkBracel)

	defer func() {
		p.expect(tkSemi)
		p.Enum = append(p.Enum, enum)
	}()

	var it int32
	for {
		p.next()
		switch p.t.T {
		case tkBracer:

			return
		case tkName:
			k := p.t.S.S
			p.next()
			switch p.t.T {
			case tkComma:
				m := EnumMember{Key: k, Value: it}
				enum.Mb = append(enum.Mb, m)
				it++
			case tkBracer:
				m := EnumMember{Key: k, Value: it}
				enum.Mb = append(enum.Mb, m)
				return
			case tkEq:
				p.expect(tkInteger)
				it = int32(p.t.S.I)
				m := EnumMember{Key: k, Value: it}
				enum.Mb = append(enum.Mb, m)
				it++
				p.next()
				if p.t.T == tkBracer {
					return
				} else if p.t.T == tkComma {
				} else {
					p.parseErr("expect , or }")
				}
			}
		}
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
	if p.t.T != tkEq {
		p.parseErr("expect ; or =")
	}
	if p.t.T == tkTMap || p.t.T == tkTVector || p.t.T == tkName {
		p.parseErr("map, vector, custom type cannot set default value")
	}

	// default
	p.next()
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
	st.TName = p.t.S.S
	for _, v := range p.Struct {
		if v.TName == st.TName {
			p.parseErr(st.TName + " Redefine.")
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
	itf.TName = p.t.S.S
	for _, v := range p.Interface {
		if v.TName == itf.TName {
			p.parseErr(itf.TName + " Redefine.")
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
		p.parseErr("const no suppost type vector or map.")
	case tkTBool, tkTByte, tkTShort,
		tkTInt, tkTLong, tkTFloat,
		tkTDouble, tkTString, tkUnsigned:
		m.Type = p.parseType()
	default:
		p.parseErr("expect type.")
	}

	p.expect(tkName)
	m.Key = p.t.S.S

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
func (p *Parse) findTNameType(tname string) (TK, string) {
	for _, v := range p.Struct {
		if p.Module+"::"+v.TName == tname {
			return tkStruct, p.Module
		}
	}

	for _, v := range p.Enum {
		if p.Module+"::"+v.TName == tname {
			return tkEnum, p.Module
		}
	}

	for _, pInc := range p.IncParse {
		ret, mod := pInc.findTNameType(tname)
		if ret != tkName {
			return ret, mod
		}
	}
	// not find
	return tkName, p.Module
}

func (p *Parse) findEnumName(ename string) (*EnumMember, *EnumInfo) {
	if strings.Contains(ename, "::") {
		ename = strings.Split(ename, "::")[1]
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
				p.parseErr(ename + " name conflict [" + cenum.TName + "::" + cmb.Key + " or " + enum.TName + "::" + mb.Key)
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
	return cmb, cenum
}

func addToSet(m *map[string]bool, module string) {
	if *m == nil {
		*m = make(map[string]bool)
	}
	(*m)[module] = true
}

func (p *Parse) checkDepTName(ty *VarType, dm *map[string]bool) {
	if ty.Type == tkName {
		name := ty.TypeSt
		if strings.Count(name, "::") == 0 {
			name = p.Module + "::" + name
		}

		mod := ""
		ty.CType, mod = p.findTNameType(name)
		if ty.CType == tkName {
			p.parseErr(ty.TypeSt + " not find define")
		}
		if mod != p.Module {
			addToSet(dm, mod)
		} else {
			// the same Module ,do not add self.
			ty.TypeSt = strings.Replace(ty.TypeSt, mod+"::", "", 1)
		}
	} else if ty.Type == tkTVector {
		p.checkDepTName(ty.TypeK, dm)
	} else if ty.Type == tkTMap {
		p.checkDepTName(ty.TypeK, dm)
		p.checkDepTName(ty.TypeV, dm)
	}
}

// analysis custom type，whether have defination
func (p *Parse) analyzeTName() {
	for i, v := range p.Struct {
		for _, v := range v.Mb {
			ty := v.Type
			p.checkDepTName(ty, &p.Struct[i].DependModule)
		}
	}

	for i, v := range p.Interface {
		for _, v := range v.Fun {
			for _, v := range v.Args {
				ty := v.Type
				p.checkDepTName(ty, &p.Interface[i].DependModule)
			}
			if v.RetType != nil {
				p.checkDepTName(v.RetType, &p.Interface[i].DependModule)
			}
		}
	}
}

func (p *Parse) analyzeDefault() {
	for _, v := range p.Struct {
		for i, r := range v.Mb {
			if r.Default != "" && r.DefType == tkName {
				mb, enum := p.findEnumName(r.Default)
				if mb == nil {
					p.parseErr("can not find default value" + r.Default)
				}
				v.Mb[i].Default = enum.TName + "_" + mb.Key
			}
		}
	}
}

// TODO analysis key[]，have quoted the correct struct and member name.
func (p *Parse) analyzeHashKey() {

}

func (p *Parse) analyzeDepend() {
	for _, v := range p.Include {
		pInc := ParseFile(v)
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

func newParse(s string, b []byte) *Parse {
	p := &Parse{Source: s}

	p.lex = NewLexState(s, b)
	return p
}

//ParseFile parse a file,return grammer tree.
func ParseFile(path string) *Parse {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("file read error: " + path + ". " + err.Error())
	}

	p := newParse(path, b)
	p.parse()

	return p
}
