package main

import (
	"bytes"
	"strconv"
	"strings"
)

//EOS is byte stream terminator
const EOS = 0

//TK is a byte type.
type TK byte

const (
	tkEos     TK = iota
	tkBracel     // ({)
	tkBracer     // }
	tkSemi       //;
	tkEq         //=
	tkShl        //<
	tkShr        //>
	tkComma      //,
	tkPtl        //(
	tkPtr        //)
	tkSquarel    //[
	tkSquarer    //]
	tkInclude    //#include

	tkDummyKeywordBegin
	// keyword
	tkModule
	tkEnum
	tkStruct
	tkInterface
	tkRequire
	tkOptional
	tkConst
	tkUnsigned
	tkVoid
	tkOut
	tkKey
	tkTrue
	tkFalse
	tkDummyKeywordEnd

	tkDummyTypeBegin
	// type
	tkTInt
	tkTBool
	tkTShort
	tkTByte
	tkTLong
	tkTFloat
	tkTDouble
	tkTString
	tkTVector
	tkTMap
	tkTArray
	tkDummyTypeEnd

	tkName // variable name
	// value
	tkString
	tkInteger
	tkFloat
)

//TokenMap record token  value.
var TokenMap = [...]string{
	tkEos: "<eos>",

	tkBracel:  "{",
	tkBracer:  "}",
	tkSemi:    ";",
	tkEq:      "=",
	tkShl:     "<",
	tkShr:     ">",
	tkComma:   ",",
	tkPtl:     "(",
	tkPtr:     ")",
	tkSquarel: "[",
	tkSquarer: "]",
	tkInclude: "#include",

	// keyword
	tkModule:    "module",
	tkEnum:      "enum",
	tkStruct:    "struct",
	tkInterface: "interface",
	tkRequire:   "require",
	tkOptional:  "optional",
	tkConst:     "const",
	tkUnsigned:  "unsigned",
	tkVoid:      "void",
	tkOut:       "out",
	tkKey:       "key",
	tkTrue:      "true",
	tkFalse:     "false",

	// type
	tkTInt:    "int",
	tkTBool:   "bool",
	tkTShort:  "short",
	tkTByte:   "byte",
	tkTLong:   "long",
	tkTFloat:  "float",
	tkTDouble: "double",
	tkTString: "string",
	tkTVector: "vector",
	tkTMap:    "map",
	tkTArray:  "array",

	tkName: "<name>",
	// value
	tkString:  "<string>",
	tkInteger: "<INTEGER>",
	tkFloat:   "<FLOAT>",
}

//SemInfo is struct.
type SemInfo struct {
	I int64
	F float64
	S string
}

//Token record token information.
type Token struct {
	T    TK
	S    *SemInfo
	Line int
}

//LexState record lexical state.
type LexState struct {
	current    byte
	linenumber int

	//t         Token
	//lookahead Token

	tokenBuff bytes.Buffer
	buff      *bytes.Buffer

	source string
}

func isNewLine(b byte) bool {
	return b == '\r' || b == '\n'
}

func isNumber(b byte) bool {
	return (b >= '0' && b <= '9') || b == '-'
}

func isHexNumber(b byte) bool {
	return (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isType(t TK) bool {
	return t > tkDummyTypeBegin && t < tkDummyTypeEnd
}

func isNumberType(t TK) bool {
	switch t {
	case tkTInt, tkTBool, tkTShort, tkTByte, tkTLong, tkTFloat, tkTDouble:
		return true
	default:
		return false
	}
}

func (ls *LexState) lexErr(err string) {
	line := strconv.Itoa(ls.linenumber)
	panic(ls.source + ": " + line + ".    " + err)
}

func (ls *LexState) incLine() {
	old := ls.current
	ls.next() /* skip '\n' or '\r' */
	if isNewLine(ls.current) && ls.current != old {
		ls.next() /* skip '\n\r' or '\r\n' */
	}
	ls.linenumber++
}

func (ls *LexState) readNumber() (TK, *SemInfo) {
	hasDot := false
	isHex := false
	sem := &SemInfo{}
	for isNumber(ls.current) || ls.current == '.' || ls.current == 'x' || ls.current == 'X' ||
		(isHex && isHexNumber(ls.current)) {

		if ls.current == '.' {
			hasDot = true
		} else if ls.current == 'x' || ls.current == 'X' {
			isHex = true
		}
		ls.tokenBuff.WriteByte(ls.current)
		ls.next()
	}
	sem.S = ls.tokenBuff.String()
	if hasDot {
		f, err := strconv.ParseFloat(sem.S, 64)
		if err != nil {
			ls.lexErr(err.Error())
		}
		sem.F = f
		return tkFloat, sem
	}
	i, err := strconv.ParseInt(sem.S, 0, 64)
	if err != nil {
		ls.lexErr(err.Error())
	}
	sem.I = i
	return tkInteger, sem
}

func (ls *LexState) readIdent() (TK, *SemInfo) {
	sem := &SemInfo{}
	var last byte

	// :: Point number processing namespace
	for isLetter(ls.current) || isNumber(ls.current) || ls.current == ':' {
		if isNumber(ls.current) && last == ':' {
			ls.lexErr("the identification is illegal.")
		}
		last = ls.current
		ls.tokenBuff.WriteByte(ls.current)
		ls.next()
	}
	sem.S = ls.tokenBuff.String()
	if strings.Count(sem.S, ":") > 0 {
		if strings.Count(sem.S, "::") != 1 || strings.Count(sem.S, ":") != 2 {
			ls.lexErr("namespace qualifier::is illegal")
		}
	}

	for i := tkDummyKeywordBegin + 1; i < tkDummyKeywordEnd; i++ {
		if TokenMap[i] == sem.S {
			return i, nil
		}
	}
	for i := tkDummyTypeBegin + 1; i < tkDummyTypeEnd; i++ {
		if TokenMap[i] == sem.S {
			return i, nil
		}
	}

	return tkName, sem
}

func (ls *LexState) readSharp() (TK, *SemInfo) {
	ls.next()
	for isLetter(ls.current) {
		ls.tokenBuff.WriteByte(ls.current)
		ls.next()
	}
	if ls.tokenBuff.String() != "include" {
		ls.lexErr("not #include")
	}

	return tkInclude, nil
}

func (ls *LexState) readString() (TK, *SemInfo) {

	sem := &SemInfo{}
	ls.next()
	for {
		if ls.current == EOS {
			ls.lexErr(`no match "`)
		} else if ls.current == '"' {
			ls.next()
			break
		} else {
			ls.tokenBuff.WriteByte(ls.current)
			ls.next()
		}
	}
	sem.S = ls.tokenBuff.String()

	return tkString, sem
}

func (ls *LexState) readLongComment() {
	for {
		switch ls.current {
		case EOS:
			ls.lexErr("respect */")
			return
		case '\n', '\r':
			ls.incLine()
		case '*':
			ls.next()
			if ls.current == EOS {
				return
			} else if ls.current == '/' {
				ls.next()
				return
			}
		default:
			ls.next()
		}
	}
}

func (ls *LexState) next() {
	var err error
	ls.current, err = ls.buff.ReadByte()
	if err != nil {
		ls.current = EOS
	}
}

func (ls *LexState) llexDefault() (TK, *SemInfo) {
	switch {
	case isNumber(ls.current):
		return ls.readNumber()
	case isLetter(ls.current):
		return ls.readIdent()
	default:
		ls.lexErr("unrecognized characters, " + string(ls.current))
		return '0', nil
	}
}

// Do lexical analysis.
func (ls *LexState) llex() (TK, *SemInfo) {
	for {
		ls.tokenBuff.Reset()
		switch ls.current {
		case EOS:
			return tkEos, nil
		case ' ', '\t', '\f', '\v':
			ls.next()
		case '\n', '\r':
			ls.incLine()
		case '/': // Comment processing
			ls.next()
			if ls.current == '/' {
				for !isNewLine(ls.current) && ls.current != EOS {
					ls.next()
				}
			} else if ls.current == '*' {
				ls.next()
				ls.readLongComment()
			} else {
				ls.lexErr("lexical error，/")
			}
		case '{':
			ls.next()
			return tkBracel, nil
		case '}':
			ls.next()
			return tkBracer, nil
		case ';':
			ls.next()
			return tkSemi, nil
		case '=':
			ls.next()
			return tkEq, nil
		case '<':
			ls.next()
			return tkShl, nil
		case '>':
			ls.next()
			return tkShr, nil
		case ',':
			ls.next()
			return tkComma, nil
		case '(':
			ls.next()
			return tkPtl, nil
		case ')':
			ls.next()
			return tkPtr, nil
		case '[':
			ls.next()
			return tkSquarel, nil
		case ']':
			ls.next()
			return tkSquarer, nil
		case '"':
			return ls.readString()
		case '#':
			return ls.readSharp()
		default:
			return ls.llexDefault()

		}
	}
}

//NextToken return token after lexical analysis.
func (ls *LexState) NextToken() *Token {
	tk := &Token{}
	tk.T, tk.S = ls.llex()
	tk.Line = ls.linenumber
	return tk
}

//NewLexState to update LexState struct.
func NewLexState(source string, buff []byte) *LexState {
	return &LexState{
		current:    ' ',
		linenumber: 1,
		source:     source,
		buff:       bytes.NewBuffer(buff),
	}
}
