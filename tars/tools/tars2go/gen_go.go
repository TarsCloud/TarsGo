package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var gE = flag.Bool("E", false, "Generate code before fmt for troubleshooting")
var gAddServant = flag.Bool("add-servant", true, "Generate AddServant function")
var gModuleCycle = flag.Bool("module-cycle", false, "support jce module cycle include(do not support jce file cycle include)")
var gModuleUpper = flag.Bool("module-upper", false, "native module names are supported, otherwise the system will upper the first letter of the module name")
var gJsonOmitEmpty = flag.Bool("json-omitempty", false, "Generate json emitempty support")
var dispatchReporter = flag.Bool("dispatch-reporter", false, "Dispatch reporter support")

var gFileMap map[string]bool

func init() {
	gFileMap = make(map[string]bool)
}

//GenGo record go code information.
type GenGo struct {
	I        []string // imports with path
	code     bytes.Buffer
	vc       int // var count. Used to generate unique variable names
	path     string
	tarsPath string
	module   string
	prefix   string
	p        *Parse

	// proto file name(not include .tars)
	ProtoName string
}

//NewGenGo build up a new path
func NewGenGo(path string, module string, outdir string) *GenGo {
	if outdir != "" {
		b := []byte(outdir)
		last := b[len(b)-1:]
		if string(last) != "/" {
			outdir += "/"
		}
	}

	return &GenGo{path: path, module: module, prefix: outdir, ProtoName: path2ProtoName(path)}
}

func path2ProtoName(path string) string {
	iBegin := strings.LastIndex(path, "/")
	if iBegin == -1 || iBegin >= len(path)-1 {
		iBegin = 0
	} else {
		iBegin++
	}
	iEnd := strings.LastIndex(path, ".tars")
	if iEnd == -1 {
		iEnd = len(path)
	}

	return path[iBegin:iEnd]
}

//Initial capitalization
func upperFirstLetter(s string) string {
	if len(s) == 0 {
		return ""
	}
	if len(s) == 1 {
		return strings.ToUpper(string(s[0]))
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}

func getShortTypeName(src string) string {
	vec := strings.Split(src, "::")
	return vec[len(vec)-1]
}

func errString(hasRet bool) string {
	var retStr string
	if hasRet {
		retStr = "return ret, err"
	} else {
		retStr = "return err"
	}
	return `if err != nil {
  ` + retStr + `
  }` + "\n"
}

func genForHead(vc string) string {
	i := `i` + vc
	e := `e` + vc
	return ` for ` + i + `,` + e + ` := int32(0),length;` + i + `<` + e + `;` + i + `++ `
}

// === rename area ===
// 0. rename module
func (p *Parse) rename() {
	p.OriginModule = p.Module
	if *gModuleUpper {
		p.Module = upperFirstLetter(p.Module)
	}
}

// 1. struct rename
// struct Name { 1 require Mb type}
func (st *StructInfo) rename() {
	st.OriginName = st.Name
	st.Name = upperFirstLetter(st.Name)
	for i := range st.Mb {
		st.Mb[i].OriginKey = st.Mb[i].Key
		st.Mb[i].Key = upperFirstLetter(st.Mb[i].Key)
	}
}

// 1. interface rename
// interface Name { Fun }
func (itf *InterfaceInfo) rename() {
	itf.OriginName = itf.Name
	itf.Name = upperFirstLetter(itf.Name)
	for i := range itf.Fun {
		itf.Fun[i].rename()
	}
}

func (en *EnumInfo) rename() {
	en.OriginName = en.Name
	en.Name = upperFirstLetter(en.Name)
	for i := range en.Mb {
		en.Mb[i].Key = upperFirstLetter(en.Mb[i].Key)
	}
}

func (cst *ConstInfo) rename() {
	cst.OriginName = cst.Name
	cst.Name = upperFirstLetter(cst.Name)
}

// 2. func rename
// type Fun (arg ArgType), in case keyword and name conflicts,argname need to capitalize.
// Fun (type int32)
func (fun *FunInfo) rename() {
	fun.OriginName = fun.Name
	fun.Name = upperFirstLetter(fun.Name)
	for i := range fun.Args {
		fun.Args[i].OriginName = fun.Args[i].Name
		// func args donot upper firs
		//fun.Args[i].Name = upperFirstLetter(fun.Args[i].Name)
	}
}

// 3. genType rename all Type

// === rename end ===

//Gen to parse file.
func (gen *GenGo) Gen() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
			// set exit code
			os.Exit(1)
		}
	}()

	gen.p = ParseFile(gen.path, make([]string, 0))
	gen.genAll()
}

func (gen *GenGo) genAll() {
	if gFileMap[gen.path] {
		// already compiled
		return
	}

	gen.p.rename()
	gen.genInclude(gen.p.IncParse)

	gen.code.Reset()
	gen.genHead()
	gen.genPackage()

	for _, v := range gen.p.Enum {
		gen.genEnum(&v)
	}

	gen.genConst(gen.p.Const)

	for _, v := range gen.p.Struct {
		gen.genStruct(&v)
	}
	if len(gen.p.Enum) > 0 || len(gen.p.Const) > 0 || len(gen.p.Struct) > 0 {
		gen.saveToSourceFile(path2ProtoName(gen.path) + ".go")
	}

	for _, v := range gen.p.Interface {
		gen.genInterface(&v)
	}

	gFileMap[gen.path] = true
}

func (gen *GenGo) genErr(err string) {
	panic(err)
}

func (gen *GenGo) saveToSourceFile(filename string) {
	var beauty []byte
	var err error
	prefix := gen.prefix

	if !*gE {
		beauty, err = format.Source(gen.code.Bytes())
		if err != nil {
			gen.genErr("go fmt fail. " + filename + " " + err.Error())
		}
	} else {
		beauty = gen.code.Bytes()
	}

	if filename == "stdout" {
		fmt.Println(string(beauty))
	} else {
		var mkPath string
		if *gModuleCycle == true {
			mkPath = prefix + gen.ProtoName + "/" + gen.p.Module
		} else {
			mkPath = prefix + gen.p.Module
		}
		err = os.MkdirAll(mkPath, 0766)

		if err != nil {
			gen.genErr(err.Error())
		}
		err = ioutil.WriteFile(mkPath+"/"+filename, beauty, 0666)

		if err != nil {
			gen.genErr(err.Error())
		}
	}
}

func (gen *GenGo) genHead() {
	gen.code.WriteString(`// Package ` + gen.p.Module + ` comment
// This file was generated by tars2go ` + VERSION + `
// Generated from ` + filepath.Base(gen.path) + `
`)
}

func (gen *GenGo) genPackage() {
	gen.code.WriteString("package " + gen.p.Module + "\n\n")
	gen.code.WriteString(`
import (
	"fmt"

`)
	gen.code.WriteString("\"" + gen.tarsPath + "/protocol/codec\"\n")

	mImports := make(map[string]bool)
	for _, st := range gen.p.Struct {
		if *gModuleCycle == true {
			for k, v := range st.DependModuleWithJce {
				gen.genStructImport(k, v, mImports)
			}
		} else {
			for k := range st.DependModule {
				gen.genStructImport(k, "", mImports)
			}
		}
	}
	for path := range mImports {
		gen.code.WriteString(path + "\n")
	}

	gen.code.WriteString(`)

	// Reference imports to suppress errors if they are not otherwise used.
	var _ = fmt.Errorf
	var _ = codec.FromInt8

`)
}

func (gen *GenGo) genStructImport(module string, protoName string, mImports map[string]bool) {
	var moduleStr string
	var jcePath string
	var moduleAlia string
	if *gModuleCycle == true {
		moduleStr = module[len(protoName)+1:]
		jcePath = protoName + "/"
		moduleAlia = module + " "
	} else {
		moduleStr = module
	}

	for _, p := range gen.I {
		if strings.HasSuffix(p, "/"+moduleStr) {
			mImports[`"`+p+`"`] = true
			return
		}
	}

	if *gModuleUpper {
		moduleAlia = upperFirstLetter(moduleAlia)
	}

	// example:
	// TarsTest.tars, MyApp
	// gomod:
	// github.com/xxx/yyy/tars-protocol/MyApp
	// github.com/xxx/yyy/tars-protocol/TarsTest/MyApp
	//
	// gopath:
	// MyApp
	// TarsTest/MyApp
	var modulePath string
	if gen.module != "" {
		mf := filepath.Clean(filepath.Join(gen.module, gen.prefix))
		modulePath = fmt.Sprintf("%s/%s%s", mf, jcePath, moduleStr)
	} else {
		modulePath = fmt.Sprintf("%s%s", jcePath, moduleStr)
	}
	mImports[moduleAlia+`"`+modulePath+`"`] = true
}

func (gen *GenGo) genIFPackage(itf *InterfaceInfo) {
	gen.code.WriteString("package " + gen.p.Module + "\n\n")
	gen.code.WriteString(`
import (
	"bytes"
	"context"
	"fmt"
	"unsafe"
	"encoding/json"
`)
	if *gAddServant {
		gen.code.WriteString("\"" + gen.tarsPath + "\"\n")
	}

	gen.code.WriteString("\"" + gen.tarsPath + "/protocol/res/requestf\"\n")
	gen.code.WriteString("m \"" + gen.tarsPath + "/model\"\n")
	gen.code.WriteString("\"" + gen.tarsPath + "/protocol/codec\"\n")
	gen.code.WriteString("\"" + gen.tarsPath + "/protocol/tup\"\n")
	gen.code.WriteString("\"" + gen.tarsPath + "/protocol/res/basef\"\n")
	gen.code.WriteString("\"" + gen.tarsPath + "/util/tools\"\n")
	gen.code.WriteString("\"" + gen.tarsPath + "/util/current\"\n")

	if *gModuleCycle == true {
		for k, v := range itf.DependModuleWithJce {
			gen.genIFImport(k, v)
		}
	} else {
		for k := range itf.DependModule {
			gen.genIFImport(k, "")
		}
	}
	gen.code.WriteString(`)

	// Reference imports to suppress errors if they are not otherwise used.
	var _ = fmt.Errorf
	var _ = codec.FromInt8
	var _ = unsafe.Pointer(nil)
	var _ = bytes.ErrTooLarge
`)
}

func (gen *GenGo) genIFImport(module string, protoName string) {
	var moduleStr string
	var jcePath string
	var moduleAlia string
	if *gModuleCycle == true {
		moduleStr = module[len(protoName)+1:]
		jcePath = protoName + "/"
		moduleAlia = module + " "
	} else {
		moduleStr = module
	}
	for _, p := range gen.I {
		if strings.HasSuffix(p, "/"+moduleStr) {
			gen.code.WriteString(`"` + p + `"` + "\n")
			return
		}
	}

	if *gModuleUpper {
		moduleAlia = upperFirstLetter(moduleAlia)
	}

	// example:
	// TarsTest.tars, MyApp
	// gomod:
	// github.com/xxx/yyy/tars-protocol/MyApp
	// github.com/xxx/yyy/tars-protocol/TarsTest/MyApp
	//
	// gopath:
	// MyApp
	// TarsTest/MyApp
	var modulePath string
	if gen.module != "" {
		mf := filepath.Clean(filepath.Join(gen.module, gen.prefix))
		modulePath = fmt.Sprintf("%s/%s%s", mf, jcePath, moduleStr)
	} else {
		modulePath = fmt.Sprintf("%s%s", jcePath, moduleStr)
	}
	gen.code.WriteString(moduleAlia + `"` + modulePath + `"` + "\n")
}

func (gen *GenGo) genType(ty *VarType) string {
	ret := ""
	switch ty.Type {
	case tkTBool:
		ret = "bool"
	case tkTInt:
		if ty.Unsigned {
			ret = "uint32"
		} else {
			ret = "int32"
		}
	case tkTShort:
		if ty.Unsigned {
			ret = "uint16"
		} else {
			ret = "int16"
		}
	case tkTByte:
		if ty.Unsigned {
			ret = "uint8"
		} else {
			ret = "int8"
		}
	case tkTLong:
		if ty.Unsigned {
			ret = "uint64"
		} else {
			ret = "int64"
		}
	case tkTFloat:
		ret = "float32"
	case tkTDouble:
		ret = "float64"
	case tkTString:
		ret = "string"
	case tkTVector:
		ret = "[]" + gen.genType(ty.TypeK)
	case tkTMap:
		ret = "map[" + gen.genType(ty.TypeK) + "]" + gen.genType(ty.TypeV)
	case tkName:
		ret = strings.Replace(ty.TypeSt, "::", ".", -1)
		vec := strings.Split(ty.TypeSt, "::")
		for i := range vec {
			if *gModuleUpper {
				vec[i] = upperFirstLetter(vec[i])
			} else {
				if i == (len(vec) - 1) {
					vec[i] = upperFirstLetter(vec[i])
				}
			}
		}
		ret = strings.Join(vec, ".")
	case tkTArray:
		ret = "[" + fmt.Sprintf("%v", ty.TypeL) + "]" + gen.genType(ty.TypeK)
	default:
		gen.genErr("Unknow Type " + TokenMap[ty.Type])
	}
	return ret
}

func (gen *GenGo) genStructDefine(st *StructInfo) {
	c := &gen.code
	c.WriteString("// " + st.Name + " struct implement\n")
	c.WriteString("type " + st.Name + " struct {\n")

	for _, v := range st.Mb {
		if *gJsonOmitEmpty {
			c.WriteString("\t" + v.Key + " " + gen.genType(v.Type) + " `json:\"" + v.OriginKey + ",omitempty\"`\n")
		} else {
			c.WriteString("\t" + v.Key + " " + gen.genType(v.Type) + " `json:\"" + v.OriginKey + "\"`\n")
		}
	}
	c.WriteString("}\n")
}

func (gen *GenGo) genFunResetDefault(st *StructInfo) {
	c := &gen.code

	c.WriteString("func (st *" + st.Name + ") ResetDefault() {\n")

	for _, v := range st.Mb {
		if v.Type.CType == tkStruct {
			c.WriteString("st." + v.Key + ".ResetDefault()\n")
		}
		if v.Default == "" {
			continue
		}
		c.WriteString("st." + v.Key + " = " + v.Default + "\n")
	}
	c.WriteString("}\n")
}

func (gen *GenGo) genWriteSimpleList(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	unsign := ""
	if mb.Type.TypeK.Unsigned {
		unsign = "u"
	}
	errStr := errString(hasRet)
	c.WriteString(`
err = _os.WriteHead(codec.SIMPLE_LIST, ` + tag + `)
` + errStr + `
err = _os.WriteHead(codec.BYTE, 0)
` + errStr + `
err = _os.Write_int32(int32(len(` + prefix + mb.Key + `)), 0)
` + errStr + `
err = _os.Write_slice_` + unsign + `int8(` + prefix + mb.Key + `)
` + errStr + `
`)
}

func (gen *GenGo) genWriteVector(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	// SIMPLE_LIST
	if mb.Type.TypeK.Type == tkTByte && !mb.Type.TypeK.Unsigned {
		gen.genWriteSimpleList(mb, prefix, hasRet)
		return
	}
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	c.WriteString(`
err = _os.WriteHead(codec.LIST, ` + tag + `)
` + errStr + `
err = _os.Write_int32(int32(len(` + prefix + mb.Key + `)), 0)
` + errStr + `
for _, v := range ` + prefix + mb.Key + ` {
`)
	// for _, v := range can nesting for _, v := range，does not conflict, support multidimensional arrays

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "v"
	gen.genWriteVar(dummy, "", hasRet)

	c.WriteString("}\n")
}

func (gen *GenGo) genWriteArray(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	// SIMPLE_LIST
	if mb.Type.TypeK.Type == tkTByte && !mb.Type.TypeK.Unsigned {
		gen.genWriteSimpleList(mb, prefix, hasRet)
		return
	}
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	c.WriteString(`
err = _os.WriteHead(codec.LIST, ` + tag + `)
` + errStr + `
err = _os.Write_int32(int32(len(` + prefix + mb.Key + `)), 0)
` + errStr + `
for _, v := range ` + prefix + mb.Key + ` {
`)
	// for _, v := range can nesting for _, v := range，does not conflict, support multidimensional arrays

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "v"
	gen.genWriteVar(dummy, "", hasRet)

	c.WriteString("}\n")
}

func (gen *GenGo) genWriteStruct(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	c.WriteString(`
err = ` + prefix + mb.Key + `.WriteBlock(_os, ` + tag + `)
` + errString(hasRet) + `
`)
}

func (gen *GenGo) genWriteMap(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	errStr := errString(hasRet)
	c.WriteString(`
err = _os.WriteHead(codec.MAP, ` + tag + `)
` + errStr + `
err = _os.Write_int32(int32(len(` + prefix + mb.Key + `)), 0)
` + errStr + `
for k` + vc + `, v` + vc + ` := range ` + prefix + mb.Key + ` {
`)
	// for _, v := range can nesting for _, v := range，does not conflict, support multidimensional arrays

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "k" + vc
	gen.genWriteVar(dummy, "", hasRet)

	dummy = &StructMember{}
	dummy.Type = mb.Type.TypeV
	dummy.Key = "v" + vc
	dummy.Tag = 1
	gen.genWriteVar(dummy, "", hasRet)

	c.WriteString("}\n")
}

func (gen *GenGo) genWriteVar(v *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	switch v.Type.Type {
	case tkTVector:
		gen.genWriteVector(v, prefix, hasRet)
	case tkTArray:
		gen.genWriteArray(v, prefix, hasRet)
	case tkTMap:
		gen.genWriteMap(v, prefix, hasRet)
	case tkName:
		if v.Type.CType == tkEnum {
			// tkEnum enumeration processing
			tag := strconv.Itoa(int(v.Tag))
			c.WriteString(`
err = _os.Write_int32(int32(` + prefix + v.Key + `),` + tag + `)
` + errString(hasRet) + `
`)
		} else {
			gen.genWriteStruct(v, prefix, hasRet)
		}
	default:
		tag := strconv.Itoa(int(v.Tag))
		c.WriteString(`
err = _os.Write_` + gen.genType(v.Type) + `(` + prefix + v.Key + `, ` + tag + `)
` + errString(hasRet) + `
`)
	}
}

func (gen *GenGo) genFunWriteBlock(st *StructInfo) {
	c := &gen.code

	// WriteBlock function head
	c.WriteString(`//WriteBlock encode struct
func (st *` + st.Name + `) WriteBlock(_os *codec.Buffer, tag byte) error {
	var err error
	err = _os.WriteHead(codec.STRUCT_BEGIN, tag)
	if err != nil {
		return err
	}

	err = st.WriteTo(_os)
	if err != nil {
		return err
	}

	err = _os.WriteHead(codec.STRUCT_END, 0)
	if err != nil {
		return err
	}
	return nil
}
`)
}

func (gen *GenGo) genFunWriteTo(st *StructInfo) {
	c := &gen.code

	c.WriteString(`//WriteTo encode struct to buffer
func (st *` + st.Name + `) WriteTo(_os *codec.Buffer) error {
	var err error
`)
	for _, v := range st.Mb {
		gen.genWriteVar(&v, "st.", false)
	}

	c.WriteString(`
	_ = err

	return nil
}
`)
}

func (gen *GenGo) genReadSimpleList(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	unsign := ""
	if mb.Type.TypeK.Unsigned {
		unsign = "u"
	}
	errStr := errString(hasRet)

	c.WriteString(`
err, _ = _is.SkipTo(codec.BYTE, 0, true)
` + errStr + `
err = _is.Read_int32(&length, 0, true)
` + errStr + `
err = _is.Read_slice_` + unsign + `int8(&` + prefix + mb.Key + `, length, true)
` + errStr + `
`)
}

func (gen *GenGo) genReadVector(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	require := "false"
	if mb.Require {
		require = "true"
	}
	c.WriteString(`
err, have, ty = _is.SkipToNoCheck(` + tag + `,` + require + `)
` + errStr + `
`)
	if require == "false" {
		c.WriteString("if have {")
	}

	c.WriteString(`
if ty == codec.LIST {
	err = _is.Read_int32(&length, 0, true)
  ` + errStr + `
  ` + prefix + mb.Key + ` = make(` + gen.genType(mb.Type) + `, length)
  ` + genForHead(vc) + `{
`)

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = mb.Key + "[i" + vc + "]"
	gen.genReadVar(dummy, prefix, hasRet)

	c.WriteString(`}
} else if ty == codec.SIMPLE_LIST {
`)
	if mb.Type.TypeK.Type == tkTByte {
		gen.genReadSimpleList(mb, prefix, hasRet)
	} else {
		c.WriteString(`err = fmt.Errorf("not support simple_list type")
    ` + errStr)
	}
	c.WriteString(`
} else {
  err = fmt.Errorf("require vector, but not")
  ` + errStr + `
}
`)

	if require == "false" {
		c.WriteString("}\n")
	}
}

func (gen *GenGo) genReadArray(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	errStr := errString(hasRet)

	// LIST
	tag := strconv.Itoa(int(mb.Tag))
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	require := "false"
	if mb.Require {
		require = "true"
	}
	c.WriteString(`
err, have, ty = _is.SkipToNoCheck(` + tag + `,` + require + `)
` + errStr + `
`)
	if require == "false" {
		c.WriteString("if have {")
	}

	c.WriteString(`
if ty == codec.LIST {
	err = _is.Read_int32(&length, 0, true)
  ` + errStr + `
  ` + genForHead(vc) + `{
`)

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = mb.Key + "[i" + vc + "]"
	gen.genReadVar(dummy, prefix, hasRet)

	c.WriteString(`}
} else if ty == codec.SIMPLE_LIST {
`)
	if mb.Type.TypeK.Type == tkTByte {
		gen.genReadSimpleList(mb, prefix, hasRet)
	} else {
		c.WriteString(`err = fmt.Errorf("not support simple_list type")
    ` + errStr)
	}
	c.WriteString(`
} else {
  err = fmt.Errorf("require array, but not")
  ` + errStr + `
}
`)

	if require == "false" {
		c.WriteString("}\n")
	}
}

func (gen *GenGo) genReadStruct(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	require := "false"
	if mb.Require {
		require = "true"
	}
	c.WriteString(`
err = ` + prefix + mb.Key + `.ReadBlock(_is, ` + tag + `, ` + require + `)
` + errString(hasRet) + `
`)
}

func (gen *GenGo) genReadMap(mb *StructMember, prefix string, hasRet bool) {
	c := &gen.code
	tag := strconv.Itoa(int(mb.Tag))
	errStr := errString(hasRet)
	vc := strconv.Itoa(gen.vc)
	gen.vc++
	require := "false"
	if mb.Require {
		require = "true"
	}
	c.WriteString(`
err, have = _is.SkipTo(codec.MAP, ` + tag + `, ` + require + `)
` + errStr + `
`)
	if require == "false" {
		c.WriteString("if have {")
	}
	c.WriteString(`
err = _is.Read_int32(&length, 0, true)
` + errStr + `
` + prefix + mb.Key + ` = make(` + gen.genType(mb.Type) + `)
` + genForHead(vc) + `{
	var k` + vc + ` ` + gen.genType(mb.Type.TypeK) + `
	var v` + vc + ` ` + gen.genType(mb.Type.TypeV) + `
`)

	dummy := &StructMember{}
	dummy.Type = mb.Type.TypeK
	dummy.Key = "k" + vc
	gen.genReadVar(dummy, "", hasRet)

	dummy = &StructMember{}
	dummy.Type = mb.Type.TypeV
	dummy.Key = "v" + vc
	dummy.Tag = 1
	gen.genReadVar(dummy, "", hasRet)

	c.WriteString(`
	` + prefix + mb.Key + `[k` + vc + `] = v` + vc + `
}
`)
	if require == "false" {
		c.WriteString("}\n")
	}
}

func (gen *GenGo) genReadVar(v *StructMember, prefix string, hasRet bool) {
	c := &gen.code

	switch v.Type.Type {
	case tkTVector:
		gen.genReadVector(v, prefix, hasRet)
	case tkTArray:
		gen.genReadArray(v, prefix, hasRet)
	case tkTMap:
		gen.genReadMap(v, prefix, hasRet)
	case tkName:
		if v.Type.CType == tkEnum {
			tag := strconv.Itoa(int(v.Tag))
			require := "false"
			if v.Require {
				require = "true"
			}
			c.WriteString(`
err = _is.Read_int32((*int32)(&` + prefix + v.Key + `),` + tag + `, ` + require + `)
` + errString(hasRet) + `
`)
		} else {
			gen.genReadStruct(v, prefix, hasRet)
		}
	default:
		tag := strconv.Itoa(int(v.Tag))
		require := "false"
		if v.Require {
			require = "true"
		}
		c.WriteString(`
err = _is.Read_` + gen.genType(v.Type) + `(&` + prefix + v.Key + `, ` + tag + `, ` + require + `)
` + errString(hasRet) + `
`)
	}
}

func (gen *GenGo) genFunReadFrom(st *StructInfo) {
	c := &gen.code

	c.WriteString(`//ReadFrom reads  from _is and put into struct.
func (st *` + st.Name + `) ReadFrom(_is *codec.Reader) error {
	var err error
	var length int32
	var have bool
	var ty byte
	st.ResetDefault()

`)

	for _, v := range st.Mb {
		gen.genReadVar(&v, "st.", false)
	}

	c.WriteString(`
	_ = err
	_ = length
	_ = have
	_ = ty
	return nil
}
`)
}

func (gen *GenGo) genFunReadBlock(st *StructInfo) {
	c := &gen.code

	c.WriteString(`//ReadBlock reads struct from the given tag , require or optional.
func (st *` + st.Name + `) ReadBlock(_is *codec.Reader, tag byte, require bool) error {
	var err error
	var have bool
	st.ResetDefault()

	err, have = _is.SkipTo(codec.STRUCT_BEGIN, tag, require)
	if err != nil {
		return err
	}
	if !have {
		if require {
			return fmt.Errorf("require ` + st.Name + `, but not exist. tag %d", tag)
		}
		return nil
	}

  	err = st.ReadFrom(_is)
  	if err != nil {
		return err
	}

	err = _is.SkipToStructEnd()
	if err != nil {
		return err
	}
	_ = have
	return nil
}
`)
}

func (gen *GenGo) genStruct(st *StructInfo) {
	gen.vc = 0
	st.rename()

	gen.genStructDefine(st)
	gen.genFunResetDefault(st)

	gen.genFunReadFrom(st)
	gen.genFunReadBlock(st)

	gen.genFunWriteTo(st)
	gen.genFunWriteBlock(st)
}

func (gen *GenGo) makeEnumName(en *EnumInfo, mb *EnumMember) string {
	return upperFirstLetter(en.Name) + "_" + upperFirstLetter(mb.Key)
}

func (gen *GenGo) genEnum(en *EnumInfo) {
	if len(en.Mb) == 0 {
		return
	}

	en.rename()

	c := &gen.code
	c.WriteString("type " + en.Name + " int32\n")
	c.WriteString("const (\n")
	var it int32
	for _, v := range en.Mb {
		if v.Type == 0 {
			//use value
			c.WriteString(gen.makeEnumName(en, &v) + ` = ` + strconv.Itoa(int(v.Value)) + "\n")
			it = v.Value + 1
		} else if v.Type == 1 {
			// use name
			find := false
			for _, ref := range en.Mb {
				if ref.Key == v.Name {
					find = true
					c.WriteString(gen.makeEnumName(en, &v) + ` = ` + gen.makeEnumName(en, &ref) + "\n")
					it = ref.Value + 1
					break
				}
				if ref.Key == v.Key {
					break
				}
			}
			if !find {
				gen.genErr(v.Name + " not define before use.")
			}
		} else {
			// use auto add
			c.WriteString(gen.makeEnumName(en, &v) + ` = ` + strconv.Itoa(int(it)) + "\n")
			it++
		}

	}

	c.WriteString(")\n")
}

func (gen *GenGo) genConst(cst []ConstInfo) {
	if len(cst) == 0 {
		return
	}

	c := &gen.code
	c.WriteString("//const as define in tars file\n")
	c.WriteString("const (\n")

	for _, v := range gen.p.Const {
		v.rename()
		c.WriteString(v.Name + " " + gen.genType(v.Type) + " = " + v.Value + "\n")
	}

	c.WriteString(")\n")
}

func (gen *GenGo) genInclude(ps []*Parse) {
	for _, v := range ps {
		gen2 := &GenGo{
			path:      v.Source,
			module:    gen.module,
			prefix:    gen.prefix,
			tarsPath:  gTarsPath,
			ProtoName: path2ProtoName(v.Source),
		}
		gen2.p = v
		gen2.genAll()
	}
}

func (gen *GenGo) genInterface(itf *InterfaceInfo) {
	gen.code.Reset()
	itf.rename()

	gen.genHead()
	gen.genIFPackage(itf)

	gen.genIFProxy(itf)

	gen.genIFServer(itf)
	gen.genIFServerWithContext(itf)

	gen.genIFDispatch(itf)

	gen.saveToSourceFile(itf.Name + ".tars.go")
}

func (gen *GenGo) genIFProxy(itf *InterfaceInfo) {
	c := &gen.code
	c.WriteString("//" + itf.Name + " struct\n")
	c.WriteString("type " + itf.Name + " struct {" + "\n")
	c.WriteString("s m.Servant" + "\n")
	c.WriteString("}" + "\n")

	for _, v := range itf.Fun {
		gen.genIFProxyFun(itf.Name, &v, false, false)
		gen.genIFProxyFun(itf.Name, &v, true, false)
		gen.genIFProxyFun(itf.Name, &v, true, true)
	}

	c.WriteString(`//SetServant sets servant for the service.
func (_obj *` + itf.Name + `) SetServant(s m.Servant) {
	_obj.s = s
}
`)
	c.WriteString(`//TarsSetTimeout sets the timeout for the servant which is in ms.
func (_obj *` + itf.Name + `) TarsSetTimeout(t int) {
	_obj.s.TarsSetTimeout(t)
}
`)

	c.WriteString(`//TarsSetProtocol sets the protocol for the servant.
func (_obj *` + itf.Name + `) TarsSetProtocol(p m.Protocol) {
	_obj.s.TarsSetProtocol(p)
}
`)

	if *gAddServant {
		c.WriteString(`//AddServant adds servant  for the service.
func (_obj *` + itf.Name + `) AddServant(imp _imp` + itf.Name + `, obj string) {
  tars.AddServant(_obj, imp, obj)
}
`)
		c.WriteString(`//AddServant adds servant  for the service with context.
func (_obj *` + itf.Name + `) AddServantWithContext(imp _imp` + itf.Name + `WithContext, obj string) {
  tars.AddServantWithContext(_obj, imp, obj)
}
`)
	}
}

func (gen *GenGo) genIFProxyFun(interfName string, fun *FunInfo, withContext bool, isOneWay bool) {
	c := &gen.code
	if withContext == true {
		if isOneWay {
			c.WriteString("//" + fun.Name + "OneWayWithContext is the proxy function for the method defined in the tars file, with the context\n")
			c.WriteString("func (_obj *" + interfName + ") " + fun.Name + "OneWayWithContext(tarsCtx context.Context,")
		} else {
			c.WriteString("//" + fun.Name + "WithContext is the proxy function for the method defined in the tars file, with the context\n")
			c.WriteString("func (_obj *" + interfName + ") " + fun.Name + "WithContext(tarsCtx context.Context,")
		}
	} else {
		c.WriteString("//" + fun.Name + " is the proxy function for the method defined in the tars file, with the context\n")
		c.WriteString("func (_obj *" + interfName + ") " + fun.Name + "(")
	}
	for _, v := range fun.Args {
		gen.genArgs(&v)
	}

	c.WriteString(" _opt ...map[string]string)")
	if fun.HasRet {
		c.WriteString("(ret " + gen.genType(fun.RetType) + ", err error){" + "\n")
	} else {
		c.WriteString("(err error)" + "{" + "\n")
	}

	c.WriteString(`
	var length int32
	var have bool
	var ty byte
  `)
	c.WriteString("_os := codec.NewBuffer()")
	var isOut bool
	for k, v := range fun.Args {
		if v.IsOut {
			isOut = true
		}
		dummy := &StructMember{}
		dummy.Type = v.Type
		dummy.Key = v.Name
		dummy.Tag = int32(k + 1)
		if v.IsOut {
			dummy.Key = "(*" + dummy.Key + ")"
		}
		gen.genWriteVar(dummy, "", fun.HasRet)
	}
	// empty args and below seperate
	c.WriteString("\n")
	errStr := errString(fun.HasRet)

	if withContext == false {
		c.WriteString(`
var _status map[string]string
var _context map[string]string
if len(_opt) == 1{
	_context =_opt[0]
}else if len(_opt) == 2 {
	_context = _opt[0]
	_status = _opt[1]
}
_resp := new(requestf.ResponsePacket)
tarsCtx := context.Background()
`)
	} else {
		c.WriteString(`var _status map[string]string
var _context map[string]string
if len(_opt) == 1{
	_context =_opt[0]
}else if len(_opt) == 2 {
	_context = _opt[0]
	_status = _opt[1]
}
_resp := new(requestf.ResponsePacket)
`)
	}

	if isOneWay {
		c.WriteString(`
		err = _obj.s.Tars_invoke(tarsCtx, 1, "` + fun.OriginName + `", _os.ToBytes(), _status, _context, _resp)
		` + errStr + `
		`)
	} else {
		c.WriteString(`
		err = _obj.s.Tars_invoke(tarsCtx, 0, "` + fun.OriginName + `", _os.ToBytes(), _status, _context, _resp)
		` + errStr + `
		`)
	}

	if (isOut || fun.HasRet) && !isOneWay {
		c.WriteString("_is := codec.NewReader(tools.Int8ToByte(_resp.SBuffer))")
	}
	if fun.HasRet && !isOneWay {
		dummy := &StructMember{}
		dummy.Type = fun.RetType
		dummy.Key = "ret"
		dummy.Tag = 0
		dummy.Require = true
		gen.genReadVar(dummy, "", fun.HasRet)
	}

	if !isOneWay {
		for k, v := range fun.Args {
			if v.IsOut {
				dummy := &StructMember{}
				dummy.Type = v.Type
				dummy.Key = "(*" + v.Name + ")"
				dummy.Tag = int32(k + 1)
				dummy.Require = true
				gen.genReadVar(dummy, "", fun.HasRet)
			}
		}
	}

	c.WriteString(`
if len(_opt) == 1{
	for k := range(_context){
		delete(_context, k)
	}
	for k, v := range(_resp.Context){
		_context[k] = v
	}
}else if len(_opt) == 2 {
	for k := range(_context){
		delete(_context, k)
	}
	for k, v := range(_resp.Context){
		_context[k] = v
	}
	for k := range(_status){
		delete(_status, k)
	}
	for k, v := range(_resp.Status){
		_status[k] = v
	}

}
  _ = length
  _ = have
  _ = ty
  `)

	if fun.HasRet {
		c.WriteString("return ret, nil" + "\n")
	} else {
		c.WriteString("return nil" + "\n")
	}

	c.WriteString("}" + "\n")
}

func (gen *GenGo) genArgs(arg *ArgInfo) {
	c := &gen.code
	c.WriteString(arg.Name + " ")
	if arg.IsOut || arg.Type.CType == tkStruct {
		c.WriteString("*")
	}

	c.WriteString(gen.genType(arg.Type) + ",")
}

func (gen *GenGo) genIFServer(itf *InterfaceInfo) {
	c := &gen.code
	c.WriteString("type _imp" + itf.Name + " interface {" + "\n")
	for _, v := range itf.Fun {
		gen.genIFServerFun(&v)
	}
	c.WriteString("}" + "\n")
}

func (gen *GenGo) genIFServerWithContext(itf *InterfaceInfo) {
	c := &gen.code
	c.WriteString("type _imp" + itf.Name + "WithContext interface {" + "\n")
	for _, v := range itf.Fun {
		gen.genIFServerFunWithContext(&v)
	}
	c.WriteString("}" + "\n")
}

func (gen *GenGo) genIFServerFun(fun *FunInfo) {
	c := &gen.code
	c.WriteString(fun.Name + "(")
	for _, v := range fun.Args {
		gen.genArgs(&v)
	}
	c.WriteString(")(")

	if fun.HasRet {
		c.WriteString("ret " + gen.genType(fun.RetType) + ", ")
	}
	c.WriteString("err error)" + "\n")
}

func (gen *GenGo) genIFServerFunWithContext(fun *FunInfo) {
	c := &gen.code
	c.WriteString(fun.Name + "(tarsCtx context.Context, ")
	for _, v := range fun.Args {
		gen.genArgs(&v)
	}
	c.WriteString(")(")

	if fun.HasRet {
		c.WriteString("ret " + gen.genType(fun.RetType) + ", ")
	}
	c.WriteString("err error)" + "\n")
}

func (gen *GenGo) genIFDispatch(itf *InterfaceInfo) {
	c := &gen.code
	c.WriteString("// Dispatch is used to call the server side implemnet for the method defined in the tars file. _withContext shows using context or not.  \n")
	c.WriteString("func(_obj *" + itf.Name + `) Dispatch(tarsCtx context.Context, _val interface{}, tarsReq *requestf.RequestPacket, tarsResp *requestf.ResponsePacket, _withContext bool) (err error) {
	var length int32
	var have bool
	var ty byte
  `)

	var param bool
	for _, v := range itf.Fun {
		if len(v.Args) > 0 {
			param = true
			break
		}
	}

	if param {
		c.WriteString("_is := codec.NewReader(tools.Int8ToByte(tarsReq.SBuffer))")
	} else {
		c.WriteString("_is := codec.NewReader(nil)")
	}
	c.WriteString(`
	_os := codec.NewBuffer()
	switch tarsReq.SFuncName {
`)

	for _, v := range itf.Fun {
		gen.genSwitchCase(itf.Name, &v)
	}

	c.WriteString(`
	default:
		return fmt.Errorf("func mismatch")
	}
	var _status map[string]string
	s, ok := current.GetResponseStatus(tarsCtx)
	if ok  && s != nil {
		_status = s
	}
	var _context map[string]string
	c, ok := current.GetResponseContext(tarsCtx)
	if ok && c != nil  {
		_context = c
	}
	*tarsResp = requestf.ResponsePacket{
		IVersion:     tarsReq.IVersion,
		CPacketType:  0,
		IRequestId:   tarsReq.IRequestId,
		IMessageType: 0,
		IRet:         0,
		SBuffer:      tools.ByteToInt8(_os.ToBytes()),
		Status:       _status,
		SResultDesc:  "",
		Context:      _context,
	}

	_ = _is
	_ = _os
	_ = length
	_ = have
	_ = ty
	return nil
}
`)
}

func (gen *GenGo) genSwitchCase(tname string, fun *FunInfo) {
	c := &gen.code
	c.WriteString(`case "` + fun.OriginName + `":` + "\n")

	inArgsCount := 0
	outArgsCount := 0
	for _, v := range fun.Args {
		c.WriteString("var " + v.Name + " " + gen.genType(v.Type) + "\n")
		if v.Type.Type == tkTMap {
			c.WriteString(v.Name + " = make(" + gen.genType(v.Type) + ")\n")
		}
		if v.IsOut {
			outArgsCount++
		} else {
			inArgsCount++
		}
	}

	//fmt.Println("args count, in, out:", inArgsCount, outArgsCount)

	c.WriteString("\n")

	if inArgsCount > 0 {
		c.WriteString("if tarsReq.IVersion == basef.TARSVERSION {" + "\n")

		for k, v := range fun.Args {
			//c.WriteString("var " + v.Name + " " + gen.genType(v.Type))
			if !v.IsOut {
				dummy := &StructMember{}
				dummy.Type = v.Type
				dummy.Key = v.Name
				dummy.Tag = int32(k + 1)
				dummy.Require = true
				gen.genReadVar(dummy, "", false)
			}
			//else {
			//	c.WriteString("\n")
			//}
		}
		//c.WriteString("}")

		c.WriteString(`} else if tarsReq.IVersion == basef.TUPVERSION {
		_reqTup_ := tup.NewUniAttribute()
		_reqTup_.Decode(_is)

		var _tupBuffer_ []byte

		`)
		for _, v := range fun.Args {
			if !v.IsOut {
				c.WriteString("\n")
				c.WriteString(`_reqTup_.GetBuffer("` + v.Name + `", &_tupBuffer_)` + "\n")
				c.WriteString("_is.Reset(_tupBuffer_)")

				dummy := &StructMember{}
				dummy.Type = v.Type
				dummy.Key = v.Name
				dummy.Tag = 0
				dummy.Require = true
				gen.genReadVar(dummy, "", false)
			}
		}

		c.WriteString(`} else if tarsReq.IVersion == basef.JSONVERSION {
		var _jsonDat_ map[string]interface{}
		_decoder_ := json.NewDecoder(bytes.NewReader(_is.ToBytes()))
		_decoder_.UseNumber()
		err = _decoder_.Decode(&_jsonDat_)
		if err != nil {
			return fmt.Errorf("Decode reqpacket failed, error: %+v", err)
		}
		`)

		for _, v := range fun.Args {
			if !v.IsOut {
				c.WriteString("{\n")
				c.WriteString(`_jsonStr_, _ := json.Marshal(_jsonDat_["` + v.Name + `"])` + "\n")
				if v.Type.CType == tkStruct {
					c.WriteString(v.Name + ".ResetDefault()\n")
				}
				c.WriteString("if err = json.Unmarshal([]byte(_jsonStr_), &" + v.Name + "); err != nil {")
				c.WriteString(`
					return err
				}
				}
				`)
			}
		}

		c.WriteString(`
		} else {
			err = fmt.Errorf("Decode reqpacket fail, error version: %d", tarsReq.IVersion)
			return err
		}`)

		c.WriteString("\n\n")
	}

	if fun.HasRet {
		c.WriteString("var _funRet_ " + gen.genType(fun.RetType) + "\n")

		c.WriteString(`if _withContext == false {
		_imp := _val.(_imp` + tname + `)
		_funRet_, err = _imp.` + fun.Name + `(`)
		for _, v := range fun.Args {
			if v.IsOut || v.Type.CType == tkStruct {
				c.WriteString("&" + v.Name + ",")
			} else {
				c.WriteString(v.Name + ",")
			}
		}
		c.WriteString(")")

		c.WriteString(`
		} else {
		_imp := _val.(_imp` + tname + `WithContext)
		_funRet_, err = _imp.` + fun.Name + `(tarsCtx ,`)
		for _, v := range fun.Args {
			if v.IsOut || v.Type.CType == tkStruct {
				c.WriteString("&" + v.Name + ",")
			} else {
				c.WriteString(v.Name + ",")
			}
		}
		c.WriteString(")" + "\n } \n")

	} else {
		c.WriteString(`if _withContext == false {
		_imp := _val.(_imp` + tname + `)
		err = _imp.` + fun.Name + `(`)
		for _, v := range fun.Args {
			if v.IsOut || v.Type.CType == tkStruct {
				c.WriteString("&" + v.Name + ",")
			} else {
				c.WriteString(v.Name + ",")
			}
		}
		c.WriteString(")")

		c.WriteString(`
		} else {
		_imp := _val.(_imp` + tname + `WithContext)
		err = _imp.` + fun.Name + `(tarsCtx ,`)
		for _, v := range fun.Args {
			if v.IsOut || v.Type.CType == tkStruct {
				c.WriteString("&" + v.Name + ",")
			} else {
				c.WriteString(v.Name + ",")
			}
		}
		c.WriteString(") \n}\n")
	}

	if *dispatchReporter {
		var inArgStr, outArgStr, retArgStr string
		if fun.HasRet {
			retArgStr = "_funRet_, err"
		} else {
			retArgStr = "err"
		}
		for _, v := range fun.Args {
			prefix := ""
			if v.Type.CType == tkStruct {
				prefix = "&"
			}
			if v.IsOut {
				outArgStr += prefix + v.Name + ","
			} else {
				inArgStr += prefix + v.Name + ","
			}
		}
		c.WriteString(`if _dp_ := tars.GetDispatchReporter(); _dp_ != nil {
			_dp_(tarsCtx, []interface{}{` + inArgStr + `}, []interface{}{` + outArgStr + `}, []interface{}{` + retArgStr + `})
		}`)

	}
	c.WriteString(`
	if err != nil {
		return err
	}
	`)

	c.WriteString(`
	if tarsReq.IVersion == basef.TARSVERSION {
	_os.Reset()
	`)

	//	if fun.HasRet {
	//		c.WriteString(`
	//		err = _os.Write_int32(_funRet_, 0)
	//		if err != nil {
	//			return err
	//		}
	//`)
	//	}

	if fun.HasRet {
		dummy := &StructMember{}
		dummy.Type = fun.RetType
		dummy.Key = "_funRet_"
		dummy.Tag = 0
		dummy.Require = true
		gen.genWriteVar(dummy, "", false)
	}

	for k, v := range fun.Args {
		if v.IsOut {
			dummy := &StructMember{}
			dummy.Type = v.Type
			dummy.Key = v.Name
			dummy.Tag = int32(k + 1)
			dummy.Require = true
			gen.genWriteVar(dummy, "", false)
		}
	}

	c.WriteString(`
} else if tarsReq.IVersion == basef.TUPVERSION {
_tupRsp_ := tup.NewUniAttribute()
`)

	//	if fun.HasRet {
	//		c.WriteString(`
	//		_os.Reset()
	//		err = _os.Write_int32(_funRet_, 0)
	//		if err != nil {
	//			return err
	//		}
	//		_tupRsp_.PutBuffer("", _os.ToBytes())
	//		_tupRsp_.PutBuffer("tars_ret", _os.ToBytes())
	//`)
	//	}

	if fun.HasRet {
		dummy := &StructMember{}
		dummy.Type = fun.RetType
		dummy.Key = "_funRet_"
		dummy.Tag = 0
		dummy.Require = true
		gen.genWriteVar(dummy, "", false)

		c.WriteString(`
		_tupRsp_.PutBuffer("", _os.ToBytes())
		_tupRsp_.PutBuffer("tars_ret", _os.ToBytes())
`)
	}

	for _, v := range fun.Args {
		if v.IsOut {
			c.WriteString(`
		_os.Reset()`)
			dummy := &StructMember{}
			dummy.Type = v.Type
			dummy.Key = v.Name
			dummy.Tag = 0
			dummy.Require = true
			gen.genWriteVar(dummy, "", false)

			c.WriteString(`_tupRsp_.PutBuffer("` + v.Name + `", _os.ToBytes())` + "\n")
		}
	}

	c.WriteString(`
	_os.Reset()
	err = _tupRsp_.Encode(_os)
	if err != nil {
		return err
	}
} else if tarsReq.IVersion == basef.JSONVERSION {
	_rspJson_ := map[string]interface{} {}
`)
	if fun.HasRet {
		//c.WriteString(`_rspJson_[""] = _funRet_` + "\n")
		c.WriteString(`_rspJson_["tars_ret"] = _funRet_` + "\n")
	}

	for _, v := range fun.Args {
		if v.IsOut {
			c.WriteString(`_rspJson_["` + v.Name + `"] = ` + v.Name + "\n")
		}
	}

	c.WriteString(`
		var _rspByte_ []byte
		if _rspByte_, err = json.Marshal(_rspJson_); err != nil {
			return err
		}

		_os.Reset()
		err = _os.Write_slice_uint8(_rspByte_)
		if err != nil {
			return err
		}
}`)

	c.WriteString("\n")

}
