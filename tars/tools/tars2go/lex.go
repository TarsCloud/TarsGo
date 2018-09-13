package main

import (
	"bytes"
	"strconv"
	"strings"
)

const EOS = 0 //字节流结束符

type TK byte

const (
	TK_EOS TK = iota

	TK_BRACEL  // {
	TK_BRACER  // }
	TK_SEMI    //;
	TK_EQ      //=
	TK_SHL     //<
	TK_SHR     //>
	TK_COMMA   //,
	TK_PTL     //(
	TK_PTR     //)
	TK_SQUAREL //[
	TK_SQUARER //]
	TK_INCLUDE //#include

	TK_DUMMY_KEYWORD_BEGIN
	// keyword
	TK_MODULE
	TK_ENUM
	TK_STRUCT
	TK_INTERFACE
	TK_REQUIRE
	TK_OPTIONAL
	TK_CONST
	TK_UNSIGNED
	TK_VOID
	TK_OUT
	TK_KEY
	TK_TRUE
	TK_FALSE
	TK_DUMMY_KEYWORD_END

	TK_DUMMY_TYPE_BEGIN
	// type
	TK_T_INT
	TK_T_BOOL
	TK_T_SHORT
	TK_T_BYTE
	TK_T_LONG
	TK_T_FLOAT
	TK_T_DOUBLE
	TK_T_STRING
	TK_T_VECTOR
	TK_T_MAP
	TK_DUMMY_TYPE_END

	TK_NAME // 变量名
	// 值
	TK_STRING
	TK_INTEGER
	TK_FLOAT
)

var TokenMap = [...]string{
	TK_EOS: "<eos>",

	TK_BRACEL:  "{",
	TK_BRACER:  "}",
	TK_SEMI:    ";",
	TK_EQ:      "=",
	TK_SHL:     "<",
	TK_SHR:     ">",
	TK_COMMA:   ",",
	TK_PTL:     "(",
	TK_PTR:     ")",
	TK_SQUAREL: "[",
	TK_SQUARER: "]",
	TK_INCLUDE: "#include",

	// keyword
	TK_MODULE:    "module",
	TK_ENUM:      "enum",
	TK_STRUCT:    "struct",
	TK_INTERFACE: "interface",
	TK_REQUIRE:   "require",
	TK_OPTIONAL:  "optional",
	TK_CONST:     "const",
	TK_UNSIGNED:  "unsigned",
	TK_VOID:      "void",
	TK_OUT:       "out",
	TK_KEY:       "key",
	TK_TRUE:      "true",
	TK_FALSE:     "false",

	// type
	TK_T_INT:    "int",
	TK_T_BOOL:   "bool",
	TK_T_SHORT:  "short",
	TK_T_BYTE:   "byte",
	TK_T_LONG:   "long",
	TK_T_FLOAT:  "float",
	TK_T_DOUBLE: "double",
	TK_T_STRING: "string",
	TK_T_VECTOR: "vector",
	TK_T_MAP:    "map",

	TK_NAME: "<name>",
	// 值
	TK_STRING:  "<string>",
	TK_INTEGER: "<INTEGER>",
	TK_FLOAT:   "<FLOAT>",
}

type SemInfo struct {
	I int64
	F float64
	S string
}

type Token struct {
	T    TK
	S    *SemInfo
	Line int
}

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
	return t > TK_DUMMY_TYPE_BEGIN && t < TK_DUMMY_TYPE_END
}

func isNumberType(t TK) bool {
	switch t {
	case TK_T_INT, TK_T_BOOL, TK_T_SHORT, TK_T_BYTE, TK_T_LONG, TK_T_FLOAT, TK_T_DOUBLE:
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
	for isNumber(ls.current) || ls.current == '.' || ls.current == 'x' || ls.current == 'X' || (isHex && isHexNumber(ls.current)) {
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
		return TK_FLOAT, sem
	} else {
		i, err := strconv.ParseInt(sem.S, 0, 64)
		if err != nil {
			ls.lexErr(err.Error())
		}
		sem.I = i
		return TK_INTEGER, sem
	}
}

func (ls *LexState) readIdent() (TK, *SemInfo) {
	sem := &SemInfo{}
	var last byte

	// :: 点号处理命名空间
	for isLetter(ls.current) || isNumber(ls.current) || ls.current == ':' {
		if isNumber(ls.current) && last == ':' {
			ls.lexErr("标识不合法")
		}
		last = ls.current
		ls.tokenBuff.WriteByte(ls.current)
		ls.next()
	}
	sem.S = ls.tokenBuff.String()
	if strings.Count(sem.S, ":") > 0 {
		if strings.Count(sem.S, "::") != 1 || strings.Count(sem.S, ":") != 2 {
			ls.lexErr("命名空间限定符::不合法")
		}
	}

	for i := TK_DUMMY_KEYWORD_BEGIN + 1; i < TK_DUMMY_KEYWORD_END; i++ {
		if TokenMap[i] == sem.S {
			return i, nil
		}
	}
	for i := TK_DUMMY_TYPE_BEGIN + 1; i < TK_DUMMY_TYPE_END; i++ {
		if TokenMap[i] == sem.S {
			return i, nil
		}
	}

	return TK_NAME, sem
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

	return TK_INCLUDE, nil
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

	return TK_STRING, sem
}

func (ls *LexState) readLongComment() {
	for {
		switch ls.current {
		case EOS:
			ls.lexErr("期待 */")
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

func (ls *LexState) llex() (TK, *SemInfo) {
	for {
		ls.tokenBuff.Reset()
		switch ls.current {
		case EOS:
			return TK_EOS, nil
		case ' ', '\t', '\f', '\v':
			ls.next()
		case '\n', '\r':
			ls.incLine()
		case '/': // 注释处理
			ls.next()
			if ls.current == '/' {
				for !isNewLine(ls.current) && ls.current != EOS {
					ls.next()
				}
			} else if ls.current == '*' {
				ls.next()
				ls.readLongComment()
			} else {
				ls.lexErr("词法错误，/")
			}
		case '{':
			ls.next()
			return TK_BRACEL, nil
		case '}':
			ls.next()
			return TK_BRACER, nil
		case ';':
			ls.next()
			return TK_SEMI, nil
		case '=':
			ls.next()
			return TK_EQ, nil
		case '<':
			ls.next()
			return TK_SHL, nil
		case '>':
			ls.next()
			return TK_SHR, nil
		case ',':
			ls.next()
			return TK_COMMA, nil
		case '(':
			ls.next()
			return TK_PTL, nil
		case ')':
			ls.next()
			return TK_PTR, nil
		case '[':
			ls.next()
			return TK_SQUAREL, nil
		case ']':
			ls.next()
			return TK_SQUARER, nil
		case '"':
			return ls.readString()
		case '#':
			return ls.readSharp()
		default:
			switch {
			case isNumber(ls.current):
				return ls.readNumber()
			case isLetter(ls.current):
				return ls.readIdent()
			default:
				ls.lexErr("不认识的字符, " + string(ls.current))
			}
		}
	}
}

func (ls *LexState) NextToken() *Token {
	tk := &Token{}
	tk.T, tk.S = ls.llex()
	tk.Line = ls.linenumber
	return tk
}

func NewLexState(source string, buff []byte) *LexState {
	return &LexState{
		current:    ' ',
		linenumber: 1,
		source:     source,
		buff:       bytes.NewBuffer(buff),
	}
}
