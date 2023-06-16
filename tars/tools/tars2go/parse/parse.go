package parse

import (
	"log"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/ast"
	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/lexer"
	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/options"
	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/token"
	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/utils"
)

// Parse record information of parse file.
type Parse struct {
	opt    *options.Options
	Module *ast.ModuleInfo

	lex    *lexer.LexState
	tk     *token.Token
	lastTk *token.Token

	// jce include chain
	IncChain            []string
	DependModuleWithJce map[string]bool

	fileNames map[string]bool
}

// NewParse parse a file,return grammar tree.
func NewParse(opt *options.Options, filePath string, incChain []string) *ast.ModuleInfo {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 查找tars文件路径
		filename := path.Base(filePath)
		for _, include := range opt.Includes {
			include = strings.TrimRight(include, "/")
			filePath = include + "/" + filename
			if _, err = os.Stat(filePath); err == nil {
				break
			}
		}
	}
	b, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalln("file read error: " + filePath + ". " + err.Error())
	}

	p := newParse(opt, filePath, b, incChain)
	p.parse()

	return p.Module
}

func newParse(opt *options.Options, source string, data []byte, incChain []string) *Parse {
	for _, v := range incChain {
		if source == v {
			panic("jce circular reference: " + source)
		}
	}
	incChain = append(incChain, source)
	log.Println(source, incChain)

	p := &Parse{
		opt: opt,
		Module: &ast.ModuleInfo{
			Source:    source,
			ProtoName: utils.Path2ProtoName(source),
		},
		lex:       lexer.NewLexState(source, data),
		IncChain:  incChain,
		fileNames: map[string]bool{},
	}
	return p
}

func (p *Parse) parseErr(err string) {
	line := "0"
	if p.tk != nil {
		line = strconv.Itoa(p.tk.Line)
	}

	panic(p.Module.Source + ": " + line + ". " + err)
}

func (p *Parse) next() {
	p.lastTk = p.tk
	p.tk = p.lex.NextToken()
}

func (p *Parse) expect(t token.Type) {
	p.next()
	if p.tk.T != t {
		p.parseErr("expect " + token.Value(t))
	}
}

func (p *Parse) makeUnsigned(utype *ast.VarType) {
	switch utype.Type {
	case token.TInt, token.TShort, token.TByte:
		utype.Unsigned = true
	default:
		p.parseErr("type " + token.Value(utype.Type) + " unsigned decoration is not supported")
	}
}

func (p *Parse) parseType() *ast.VarType {
	vtype := &ast.VarType{Type: p.tk.T}

	switch vtype.Type {
	case token.Name:
		vtype.TypeSt = p.tk.S.S
	case token.TInt, token.TBool, token.TShort, token.TLong, token.TByte, token.TFloat, token.TDouble, token.TString:
		// no nothing
	case token.TVector:
		p.expect(token.Shl)
		p.next()
		vtype.TypeK = p.parseType()
		p.expect(token.Shr)
	case token.TMap:
		p.expect(token.Shl)
		p.next()
		vtype.TypeK = p.parseType()
		p.expect(token.Comma)
		p.next()
		vtype.TypeV = p.parseType()
		p.expect(token.Shr)
	case token.Unsigned:
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
	enum := ast.EnumInfo{}
	p.expect(token.Name)
	enum.Name = p.tk.S.S
	for _, v := range p.Module.Enum {
		if v.Name == enum.Name {
			p.parseErr(enum.Name + " Redefine.")
		}
	}
	p.expect(token.BraceLeft)

LFOR:
	for {
		p.next()
		switch p.tk.T {
		case token.BraceRight:
			break LFOR
		case token.Name:
			k := p.tk.S.S
			p.next()
			switch p.tk.T {
			case token.Comma:
				m := ast.EnumMember{Key: k, Type: 2}
				enum.Mb = append(enum.Mb, m)
			case token.BraceRight:
				m := ast.EnumMember{Key: k, Type: 2}
				enum.Mb = append(enum.Mb, m)
				break LFOR
			case token.Eq:
				p.next()
				switch p.tk.T {
				case token.Integer:
					m := ast.EnumMember{Key: k, Value: int32(p.tk.S.I)}
					enum.Mb = append(enum.Mb, m)
				case token.Name:
					m := ast.EnumMember{Key: k, Type: 1, Name: p.tk.S.S}
					enum.Mb = append(enum.Mb, m)
				default:
					p.parseErr("not expect " + token.Value(p.tk.T))
				}
				p.next()
				if p.tk.T == token.BraceRight {
					break LFOR
				} else if p.tk.T == token.Comma {
				} else {
					p.parseErr("expect , or }")
				}
			}
		}
	}
	p.expect(token.Semi)
	p.Module.Enum = append(p.Module.Enum, enum)
}

func (p *Parse) parseStructMemberDefault(m *ast.StructMember) {
	m.DefType = p.tk.T
	switch p.tk.T {
	case token.Integer:
		if !token.IsNumberType(m.Type.Type) && m.Type.Type != token.Name {
			// enum auto defined type ,default value is number.
			p.parseErr("type does not accept number")
		}
		m.Default = p.tk.S.S
	case token.Float:
		if !token.IsNumberType(m.Type.Type) {
			p.parseErr("type does not accept number")
		}
		m.Default = p.tk.S.S
	case token.String:
		if token.IsNumberType(m.Type.Type) {
			p.parseErr("type does not accept string")
		}
		m.Default = `"` + p.tk.S.S + `"`
	case token.True:
		if m.Type.Type != token.TBool {
			p.parseErr("default value format error")
		}
		m.Default = "true"
	case token.False:
		if m.Type.Type != token.TBool {
			p.parseErr("default value format error")
		}
		m.Default = "false"
	case token.Name:
		m.Default = p.tk.S.S
	default:
		p.parseErr("default value format error")
	}
}

func (p *Parse) parseStructMember() *ast.StructMember {
	// tag or end
	p.next()
	if p.tk.T == token.BraceRight {
		return nil
	}
	if p.tk.T != token.Integer {
		p.parseErr("expect tags.")
	}
	m := &ast.StructMember{}
	m.Tag = int32(p.tk.S.I)

	// require or optional
	p.next()
	if p.tk.T == token.Require {
		m.Require = true
	} else if p.tk.T == token.Optional {
		m.Require = false
	} else {
		p.parseErr("expect require or optional")
	}

	// type
	p.next()
	if !token.IsType(p.tk.T) && p.tk.T != token.Name && p.tk.T != token.Unsigned {
		p.parseErr("expect type")
	} else {
		m.Type = p.parseType()
	}

	// key
	p.expect(token.Name)
	m.Key = p.tk.S.S

	p.next()
	if p.tk.T == token.Semi {
		return m
	}
	if p.tk.T == token.SquareLeft {
		p.expect(token.Integer)
		m.Type = &ast.VarType{Type: token.TArray, TypeK: m.Type, TypeL: p.tk.S.I}
		p.expect(token.SquarerRight)
		p.expect(token.Semi)
		return m
	}
	if p.tk.T != token.Eq {
		p.parseErr("expect ; or =")
	}
	if p.tk.T == token.TMap || p.tk.T == token.TVector || p.tk.T == token.Name {
		p.parseErr("map, vector, custom type cannot set default value")
	}

	// default
	p.next()
	p.parseStructMemberDefault(m)
	p.expect(token.Semi)

	return m
}

func (p *Parse) checkTag(st *ast.StructInfo) {
	set := make(map[int32]bool)
	for _, v := range st.Mb {
		if set[v.Tag] {
			p.parseErr("tag = " + strconv.Itoa(int(v.Tag)) + ". have duplicates")
		}
		set[v.Tag] = true
	}
}

func (p *Parse) sortTag(st *ast.StructInfo) {
	sort.Sort(ast.StructMemberSorter(st.Mb))
}

func (p *Parse) parseStruct() {
	st := ast.StructInfo{}
	p.expect(token.Name)
	st.Name = p.tk.S.S
	for _, v := range p.Module.Struct {
		if v.Name == st.Name {
			p.parseErr(st.Name + " Redefine.")
		}
	}
	p.expect(token.BraceLeft)

	for {
		m := p.parseStructMember()
		if m == nil {
			break
		}
		st.Mb = append(st.Mb, *m)
	}
	p.expect(token.Semi) //semicolon at the end of the struct.

	p.checkTag(&st)
	p.sortTag(&st)

	p.Module.Struct = append(p.Module.Struct, st)
}

func (p *Parse) parseInterfaceFun() *ast.FunInfo {
	fun := &ast.FunInfo{}
	p.next()
	if p.tk.T == token.BraceRight {
		return nil
	}
	if p.tk.T == token.Void {
		fun.HasRet = false
	} else if !token.IsType(p.tk.T) && p.tk.T != token.Name && p.tk.T != token.Unsigned {
		p.parseErr("expect type")
	} else {
		fun.HasRet = true
		fun.RetType = p.parseType()
	}
	p.expect(token.Name)
	fun.Name = p.tk.S.S
	p.expect(token.Ptl)

	p.next()
	if p.tk.T == token.Shr {
		return fun
	}

	// No parameter function, exit directly.
	if p.tk.T == token.Ptr {
		p.expect(token.Semi)
		return fun
	}

	for {
		arg := &ast.ArgInfo{}
		if p.tk.T == token.Out {
			arg.IsOut = true
			p.next()
		} else {
			arg.IsOut = false
		}

		arg.Type = p.parseType()
		p.next()
		if p.tk.T == token.Name {
			arg.Name = p.tk.S.S
			p.next()
		}

		fun.Args = append(fun.Args, *arg)

		if p.tk.T == token.Comma {
			p.next()
		} else if p.tk.T == token.Ptr {
			p.expect(token.Semi)
			break
		} else {
			p.parseErr("expect , or )")
		}
	}
	return fun
}

func (p *Parse) parseInterface() {
	itf := &ast.InterfaceInfo{}
	p.expect(token.Name)
	itf.Name = p.tk.S.S
	for _, v := range p.Module.Interface {
		if v.Name == itf.Name {
			p.parseErr(itf.Name + " Redefine.")
		}
	}
	p.expect(token.BraceLeft)

	for {
		fun := p.parseInterfaceFun()
		if fun == nil {
			break
		}
		itf.Fun = append(itf.Fun, *fun)
	}
	p.expect(token.Semi) //semicolon at the end of struct.
	p.Module.Interface = append(p.Module.Interface, *itf)
}

func (p *Parse) parseConst() {
	m := ast.ConstInfo{}

	// type
	p.next()
	switch p.tk.T {
	case token.TVector, token.TMap:
		p.parseErr("const no supports type vector or map.")
	case token.TBool, token.TByte, token.TShort,
		token.TInt, token.TLong, token.TFloat,
		token.TDouble, token.TString, token.Unsigned:
		m.Type = p.parseType()
	default:
		p.parseErr("expect type.")
	}

	p.expect(token.Name)
	m.Name = p.tk.S.S

	p.expect(token.Eq)

	// default
	p.next()
	switch p.tk.T {
	case token.Integer, token.Float:
		if !token.IsNumberType(m.Type.Type) {
			p.parseErr("type does not accept number")
		}
		m.Value = p.tk.S.S
	case token.String:
		if token.IsNumberType(m.Type.Type) {
			p.parseErr("type does not accept string")
		}
		m.Value = `"` + p.tk.S.S + `"`
	case token.True:
		if m.Type.Type != token.TBool {
			p.parseErr("default value format error")
		}
		m.Value = "true"
	case token.False:
		if m.Type.Type != token.TBool {
			p.parseErr("default value format error")
		}
		m.Value = "false"
	default:
		p.parseErr("default value format error")
	}
	p.expect(token.Semi)

	p.Module.Const = append(p.Module.Const, m)
}

func (p *Parse) parseHashKey() {
	hashKey := ast.HashKeyInfo{}
	p.expect(token.SquareLeft)
	p.expect(token.Name)
	hashKey.Name = p.tk.S.S
	p.expect(token.Comma)
	for {
		p.expect(token.Name)
		hashKey.Member = append(hashKey.Member, p.tk.S.S)
		p.next()
		t := p.tk
		switch t.T {
		case token.SquarerRight:
			p.expect(token.Semi)
			p.Module.HashKey = append(p.Module.HashKey, hashKey)
			return
		case token.Comma:
		default:
			p.parseErr("expect ] or ,")
		}
	}
}

func (p *Parse) parseModuleSegment() {
	p.expect(token.BraceLeft)

	for {
		p.next()
		t := p.tk
		switch t.T {
		case token.BraceRight:
			p.expect(token.Semi)
			return
		case token.Const:
			p.parseConst()
		case token.Enum:
			p.parseEnum()
		case token.Struct:
			p.parseStruct()
		case token.Interface:
			p.parseInterface()
		case token.Key:
			p.parseHashKey()
		default:
			p.parseErr("not except " + token.Value(t.T))
		}
	}
}

func (p *Parse) parseModule() {
	p.expect(token.Name)

	// 解决一个tars文件中定义多个module
	if p.Module.Name != "" {
		name := p.Module.ProtoName + "_" + p.tk.S.S + ".tars"
		newp := newParse(p.opt, p.Module.Source, nil, nil)
		newp.Module.Name = p.tk.S.S
		newp.Module.Include = p.Module.Include
		m := *p.Module
		newp.Module.IncModule = append(newp.Module.IncModule, &m)
		newp.lex = p.lex
		newp.parseModuleSegment()
		newp.analyzeDepend()
		if p.fileNames[name] {
			// merge
			for _, module := range p.Module.IncModule {
				if module.Name == newp.Module.Name {
					module.Struct = append(module.Struct, newp.Module.Struct...)
					module.Interface = append(module.Interface, newp.Module.Interface...)
					module.Enum = append(module.Enum, newp.Module.Enum...)
					module.Const = append(module.Const, newp.Module.Const...)
					module.HashKey = append(module.HashKey, newp.Module.HashKey...)
					break
				}
			}
		} else {
			// 增加已经解析的module
			p.Module.IncModule = append(p.Module.IncModule, newp.Module)
			p.fileNames[name] = true
		}
		p.lex = newp.lex
	} else {
		p.Module.Name = p.tk.S.S
		p.parseModuleSegment()
	}
}

func (p *Parse) parseInclude() {
	p.expect(token.String)
	p.Module.Include = append(p.Module.Include, p.tk.S.S)
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

func (p *Parse) checkDepTName(ty *ast.VarType, dm *map[string]bool, dmj *map[string]string) {
	if ty.Type == token.Name {
		name := ty.TypeSt
		if strings.Count(name, "::") == 0 {
			name = p.Module.Name + "::" + name
		}

		mod := ""
		protoName := ""
		ty.CType, mod, protoName = p.Module.FindTNameType(name)
		if ty.CType == token.Name {
			p.parseErr(ty.TypeSt + " not find define")
		}
		if p.opt.ModuleCycle {
			if mod != p.Module.Name || protoName != p.Module.ProtoName {
				var modStr string
				if p.opt.ModuleUpper {
					modStr = utils.UpperFirstLetter(mod)
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
			if mod != p.Module.Name {
				addToSet(dm, mod)
			} else {
				// the same Module ,do not add self.
				ty.TypeSt = strings.Replace(ty.TypeSt, mod+"::", "", 1)
			}
		}
	} else if ty.Type == token.TVector {
		p.checkDepTName(ty.TypeK, dm, dmj)
	} else if ty.Type == token.TMap {
		p.checkDepTName(ty.TypeK, dm, dmj)
		p.checkDepTName(ty.TypeV, dm, dmj)
	}
}

// analysis custom type，whether have definition
func (p *Parse) analyzeTName() {
	for i, v := range p.Module.Struct {
		for _, v := range v.Mb {
			ty := v.Type
			p.checkDepTName(ty, &p.Module.Struct[i].DependModule, &p.Module.Struct[i].DependModuleWithJce)
		}
	}

	for i, v := range p.Module.Interface {
		for _, v := range v.Fun {
			for _, v := range v.Args {
				ty := v.Type
				p.checkDepTName(ty, &p.Module.Interface[i].DependModule, &p.Module.Interface[i].DependModuleWithJce)
			}
			if v.RetType != nil {
				p.checkDepTName(v.RetType, &p.Module.Interface[i].DependModule, &p.Module.Interface[i].DependModuleWithJce)
			}
		}
	}
}

func (p *Parse) analyzeDefault() {
	for _, v := range p.Module.Struct {
		for i, r := range v.Mb {
			if r.Default != "" && r.DefType == token.Name {
				mb, enum, err := p.Module.FindEnumName(r.Default, p.opt.ModuleCycle)
				if err != nil {
					p.parseErr(err.Error())
				}
				if mb == nil || enum == nil {
					p.parseErr("can not find default value" + r.Default)
				}
				defValue := enum.Name + "_" + utils.UpperFirstLetter(mb.Key)
				var currModule string
				if p.opt.ModuleCycle {
					currModule = p.Module.ProtoName + "_" + p.Module.Name
				} else {
					currModule = p.Module.Name
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
	for _, v := range p.Module.Include {
		relativePath := path.Dir(p.Module.Source)
		dependFile := relativePath + "/" + v
		pInc := NewParse(p.opt, dependFile, p.IncChain)
		p.Module.IncModule = append(p.Module.IncModule, pInc)
		log.Println("parse include: ", v)
	}

	p.analyzeDefault()
	p.analyzeTName()
	p.analyzeHashKey()
}

func (p *Parse) parse() {
OUT:
	for {
		p.next()
		t := p.tk
		switch t.T {
		case token.Eof:
			break OUT
		case token.Include:
			p.parseInclude()
		case token.Module:
			p.parseModule()
		default:
			p.parseErr("Expect include or module.")
		}
	}
	p.analyzeDepend()
}
