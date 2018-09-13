package main

import (
	"fmt"
	"io/ioutil"
	"sort"
	"strconv"
	"strings"
)

type VarType struct {
	Type     TK       // 基础类型
	Unsigned bool     // 是否是无符号
	TypeSt   string   // 自定义类型名称，如枚举结构体，此时Type=TK_NAME
	CType    TK       // 确定自定义类型时哪一种，TK_ENUM, TK_STRUCT
	TypeK    *VarType // vector的成员变量，map的key
	TypeV    *VarType // map的value
}

type StructMember struct {
	Tag     int32
	Require bool
	Type    *VarType
	Key     string // 经过大写转换后的key
	KeyStr  string // 原始key
	Default string
	DefType TK
}

// 序列化时，要保证tag是有序的
type StructMemberSorter []StructMember

func (a StructMemberSorter) Len() int           { return len(a) }
func (a StructMemberSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a StructMemberSorter) Less(i, j int) bool { return a[i].Tag < a[j].Tag }

type StructInfo struct {
	TName        string
	Mb           []StructMember
	DependModule map[string]bool
}

type ArgInfo struct {
	Name  string
	IsOut bool
	Type  *VarType
}
type FunInfo struct {
	Name    string // 经过大写转换后的name
	NameStr string // 原始name
	HasRet  bool
	RetType *VarType
	Args    []ArgInfo
}

type InterfaceInfo struct {
	TName        string
	Fun          []FunInfo
	DependModule map[string]bool
}

type EnumMember struct {
	Key   string
	Value int32
}
type EnumInfo struct {
	TName string
	Mb    []EnumMember
}

type ConstInfo struct {
	Type  *VarType
	Key   string
	Value string
}

type HashKeyInfo struct {
	Name   string
	Member []string
}

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

	// 解析好的include文件
	IncParse []*Parse

	lex    *LexState
	t      *Token
	last_t *Token
}

func (p *Parse) parseErr(err string) {
	line := "0"
	if p.t != nil {
		line = strconv.Itoa(p.t.Line)
	}

	panic(p.Source + ": " + line + ". " + err)
}

func (p *Parse) next() {
	p.last_t = p.t
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
	case TK_T_INT, TK_T_SHORT, TK_T_BYTE:
		utype.Unsigned = true
	default:
		p.parseErr("type " + TokenMap[utype.Type] + " 不支持unsigned修饰")
	}
}

func (p *Parse) parseType() *VarType {
	vtype := &VarType{Type: p.t.T}

	switch vtype.Type {
	case TK_NAME:
		vtype.TypeSt = p.t.S.S
	case TK_T_INT, TK_T_BOOL, TK_T_SHORT, TK_T_LONG, TK_T_BYTE, TK_T_FLOAT, TK_T_DOUBLE, TK_T_STRING:
		// no nothing
	case TK_T_VECTOR:
		p.expect(TK_SHL)
		p.next()
		vtype.TypeK = p.parseType()
		p.expect(TK_SHR)
	case TK_T_MAP:
		p.expect(TK_SHL)
		p.next()
		vtype.TypeK = p.parseType()
		p.expect(TK_COMMA)
		p.next()
		vtype.TypeV = p.parseType()
		p.expect(TK_SHR)
	case TK_UNSIGNED:
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
	p.expect(TK_NAME)
	enum.TName = p.t.S.S
	for _, v := range p.Enum {
		if v.TName == enum.TName {
			p.parseErr(enum.TName + " Redefine.")
		}
	}
	p.expect(TK_BRACEL)

	defer func() {
		p.expect(TK_SEMI)
		p.Enum = append(p.Enum, enum)
	}()

	var it int32
	for {
		p.next()
		switch p.t.T {
		case TK_BRACER:
			//if p.last_t.T == TK_COMMA {
			//	p.parseErr("逗号后面不能紧跟右大括号")
			//}
			return
		case TK_NAME:
			k := p.t.S.S
			p.next()
			switch p.t.T {
			case TK_COMMA:
				m := EnumMember{Key: k, Value: it}
				enum.Mb = append(enum.Mb, m)
				it++
			case TK_BRACER:
				m := EnumMember{Key: k, Value: it}
				enum.Mb = append(enum.Mb, m)
				return
			case TK_EQ:
				p.expect(TK_INTEGER)
				it = int32(p.t.S.I)
				m := EnumMember{Key: k, Value: it}
				enum.Mb = append(enum.Mb, m)
				it++
				p.next()
				if p.t.T == TK_BRACER {
					return
				} else if p.t.T == TK_COMMA {
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
	if p.t.T == TK_BRACER {
		return nil
	}
	if p.t.T != TK_INTEGER {
		p.parseErr("expect tags.")
	}
	m := &StructMember{}
	m.Tag = int32(p.t.S.I)

	// require or optional
	p.next()
	if p.t.T == TK_REQUIRE {
		m.Require = true
	} else if p.t.T == TK_OPTIONAL {
		m.Require = false
	} else {
		p.parseErr("expect require or optional")
	}

	// type
	p.next()
	if !isType(p.t.T) && p.t.T != TK_NAME && p.t.T != TK_UNSIGNED {
		p.parseErr("expect type")
	} else {
		m.Type = p.parseType()
	}

	// key
	p.expect(TK_NAME)
	m.Key = p.t.S.S

	p.next()
	if p.t.T == TK_SEMI {
		return m
	}
	if p.t.T != TK_EQ {
		p.parseErr("expect ; or =")
	}
	if p.t.T == TK_T_MAP || p.t.T == TK_T_VECTOR || p.t.T == TK_NAME {
		p.parseErr("map, vector, 自定义类型 不能设置默认值")
	}

	// default
	p.next()
	m.DefType = p.t.T
	switch p.t.T {
	case TK_INTEGER:
		if !isNumberType(m.Type.Type) && m.Type.Type != TK_NAME {
			// enum 自定义类型 默认值 也是数字
			p.parseErr("类型不接受数字")
		}
		m.Default = p.t.S.S
	case TK_FLOAT:
		if !isNumberType(m.Type.Type) {
			p.parseErr("类型不接受数字")
		}
		m.Default = p.t.S.S
	case TK_STRING:
		if isNumberType(m.Type.Type) {
			p.parseErr("类型不接受字符串")
		}
		m.Default = `"` + p.t.S.S + `"`
	case TK_TRUE:
		if m.Type.Type != TK_T_BOOL {
			p.parseErr("默认值格式错误")
		}
		m.Default = "true"
	case TK_FALSE:
		if m.Type.Type != TK_T_BOOL {
			p.parseErr("默认值格式错误")
		}
		m.Default = "false"
	case TK_NAME:
		m.Default = p.t.S.S
	default:
		p.parseErr("默认值格式错误")
	}
	p.expect(TK_SEMI)

	return m
}

func (p *Parse) checkTag(st *StructInfo) {
	set := make(map[int32]bool)
	for _, v := range st.Mb {
		if set[v.Tag] {
			p.parseErr("tag = " + strconv.Itoa(int(v.Tag)) + ". 有重复")
		}
		set[v.Tag] = true
	}
}

func (p *Parse) sortTag(st *StructInfo) {
	sort.Sort(StructMemberSorter(st.Mb))
}

func (p *Parse) parseStruct() {
	st := StructInfo{}
	p.expect(TK_NAME)
	st.TName = p.t.S.S
	for _, v := range p.Struct {
		if v.TName == st.TName {
			p.parseErr(st.TName + " Redefine.")
		}
	}
	p.expect(TK_BRACEL)

	for {
		m := p.parseStructMember()
		if m == nil {
			break
		}
		st.Mb = append(st.Mb, *m)
	}
	p.expect(TK_SEMI) //结构体结尾的分号

	p.checkTag(&st)
	p.sortTag(&st)

	p.Struct = append(p.Struct, st)
}

func (p *Parse) parseInterfaceFun() *FunInfo {
	fun := &FunInfo{}
	p.next()
	if p.t.T == TK_BRACER {
		return nil
	}
	if p.t.T == TK_VOID {
		fun.HasRet = false
	} else if !isType(p.t.T) && p.t.T != TK_NAME && p.t.T != TK_UNSIGNED {
		p.parseErr("expect type")
	} else {
		fun.HasRet = true
		fun.RetType = p.parseType()
	}
	p.expect(TK_NAME)
	fun.Name = p.t.S.S
	p.expect(TK_PTL)

	p.next()
	if p.t.T == TK_SHR {
		return fun
	}

	// 无参函数，直接退出
	if p.t.T == TK_PTR {
		p.expect(TK_SEMI)
		return fun
	}

	for {
		arg := &ArgInfo{}
		if p.t.T == TK_OUT {
			arg.IsOut = true
			p.next()
		} else {
			arg.IsOut = false
		}

		arg.Type = p.parseType()
		p.next()
		if p.t.T == TK_NAME {
			arg.Name = p.t.S.S
			p.next()
		}

		fun.Args = append(fun.Args, *arg)

		if p.t.T == TK_COMMA {
			p.next()
		} else if p.t.T == TK_PTR {
			p.expect(TK_SEMI)
			break
		} else {
			p.parseErr("expect , or )")
		}
	}
	return fun
}

func (p *Parse) parseInterface() {
	itf := &InterfaceInfo{}
	p.expect(TK_NAME)
	itf.TName = p.t.S.S
	for _, v := range p.Interface {
		if v.TName == itf.TName {
			p.parseErr(itf.TName + " Redefine.")
		}
	}
	p.expect(TK_BRACEL)

	for {
		fun := p.parseInterfaceFun()
		if fun == nil {
			break
		}
		itf.Fun = append(itf.Fun, *fun)
	}
	p.expect(TK_SEMI) //结构体结尾的分号
	p.Interface = append(p.Interface, *itf)
}

func (p *Parse) parseConst() {
	m := ConstInfo{}

	// type
	p.next()
	switch p.t.T {
	case TK_T_VECTOR, TK_T_MAP:
		p.parseErr("const no suppost type vector or map.")
	case TK_T_BOOL, TK_T_BYTE, TK_T_SHORT,
		TK_T_INT, TK_T_LONG, TK_T_FLOAT,
		TK_T_DOUBLE, TK_T_STRING, TK_UNSIGNED:
		m.Type = p.parseType()
	default:
		p.parseErr("expect type.")
	}

	p.expect(TK_NAME)
	m.Key = p.t.S.S

	p.expect(TK_EQ)

	// default
	p.next()
	switch p.t.T {
	case TK_INTEGER, TK_FLOAT:
		if !isNumberType(m.Type.Type) {
			p.parseErr("类型不接受数字")
		}
		m.Value = p.t.S.S
	case TK_STRING:
		if isNumberType(m.Type.Type) {
			p.parseErr("类型不接受字符串")
		}
		m.Value = `"` + p.t.S.S + `"`
	case TK_TRUE:
		if m.Type.Type != TK_T_BOOL {
			p.parseErr("默认值格式错误")
		}
		m.Value = "true"
	case TK_FALSE:
		if m.Type.Type != TK_T_BOOL {
			p.parseErr("默认值格式错误")
		}
		m.Value = "false"
	default:
		p.parseErr("默认值格式错误")
	}
	p.expect(TK_SEMI)

	p.Const = append(p.Const, m)
}

func (p *Parse) parseHashKey() {
	hashKey := HashKeyInfo{}
	p.expect(TK_SQUAREL)
	p.expect(TK_NAME)
	hashKey.Name = p.t.S.S
	p.expect(TK_COMMA)
	for {
		p.expect(TK_NAME)
		hashKey.Member = append(hashKey.Member, p.t.S.S)
		p.next()
		t := p.t
		switch t.T {
		case TK_SQUARER:
			p.expect(TK_SEMI)
			p.HashKey = append(p.HashKey, hashKey)
			return
		case TK_COMMA:
		default:
			p.parseErr("expect ] or ,")
		}
	}
}

func (p *Parse) parseModuleSegment() {
	p.expect(TK_BRACEL)

	for {
		p.next()
		t := p.t
		switch t.T {
		case TK_BRACER:
			p.expect(TK_SEMI)
			return
		case TK_CONST:
			p.parseConst()
		case TK_ENUM:
			p.parseEnum()
		case TK_STRUCT:
			p.parseStruct()
		case TK_INTERFACE:
			p.parseInterface()
		case TK_KEY:
			p.parseHashKey()
		default:
			p.parseErr("not except " + TokenMap[t.T])
		}
	}
}

func (p *Parse) parseModule() {
	p.expect(TK_NAME)

	if p.Module != "" {
		p.parseErr("不要重复定义module")
	}
	p.Module = p.t.S.S

	p.parseModuleSegment()
}

func (p *Parse) parseInclude() {
	p.expect(TK_STRING)
	p.Include = append(p.Include, p.t.S.S)
}

// 寻找用户自定义标识的真正类型
func (p *Parse) findTNameType(tname string) (TK, string) {
	for _, v := range p.Struct {
		if p.Module+"::"+v.TName == tname {
			return TK_STRUCT, p.Module
		}
	}

	for _, v := range p.Enum {
		if p.Module+"::"+v.TName == tname {
			return TK_ENUM, p.Module
		}
	}

	for _, pInc := range p.IncParse {
		ret, mod := pInc.findTNameType(tname)
		if ret != TK_NAME {
			return ret, mod
		}
	}
	// not find
	return TK_NAME, p.Module
}

func (p *Parse) findEnumName(ename string) (*EnumMember, *EnumInfo) {
	if strings.Contains(ename, "::") {
		ename = strings.Split(ename, "::")[1]
	}
	var cmb *EnumMember = nil
	var cenum *EnumInfo = nil
	for ek, enum := range p.Enum {
		for mk, mb := range enum.Mb {
			if mb.Key != ename {
				continue
			}
			if cmb == nil {
				cmb = &enum.Mb[mk]
				cenum = &p.Enum[ek]
			} else {
				p.parseErr(ename + " 名冲突 [" + cenum.TName + "::" + cmb.Key + " 或 " + enum.TName + "::" + mb.Key)
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
	if ty.Type == TK_NAME {
		name := ty.TypeSt
		if strings.Count(name, "::") == 0 {
			name = p.Module + "::" + name
		}

		mod := ""
		ty.CType, mod = p.findTNameType(name)
		if ty.CType == TK_NAME {
			p.parseErr(ty.TypeSt + " not find define")
		}
		if mod != p.Module {
			addToSet(dm, mod)
		} else {
			// 同一个Module 不要加自己
			ty.TypeSt = strings.Replace(ty.TypeSt, mod+"::", "", 1)
		}
	} else if ty.Type == TK_T_VECTOR {
		p.checkDepTName(ty.TypeK, dm)
	} else if ty.Type == TK_T_MAP {
		p.checkDepTName(ty.TypeK, dm)
		p.checkDepTName(ty.TypeV, dm)
	}
}

// 分析自定义类型，是否都有定义
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
			if r.Default != "" && r.DefType == TK_NAME {
				mb, enum := p.findEnumName(r.Default)
				if mb == nil {
					p.parseErr("找不到默认值" + r.Default)
				}
				v.Mb[i].Default = enum.TName + "_" + mb.Key
			}
		}
	}
}

// TODO 分析key[]，是否都引用了正确的结构体和成员名
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
		case TK_EOS:
			break OUT
		case TK_INCLUDE:
			p.parseInclude()
		case TK_MODULE:
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

func ParseFile(path string) *Parse {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("文件读取错误: " + path + ". " + err.Error())
	}

	p := newParse(path, b)
	p.parse()

	return p
}
