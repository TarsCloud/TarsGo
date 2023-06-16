package lexer

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/TarsCloud/TarsGo/tars/tools/tars2go/token"
)

// LexState record lexical state.
type LexState struct {
	current    byte
	lineNumber int

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

func (ls *LexState) lexErr(err string) {
	line := strconv.Itoa(ls.lineNumber)
	panic(ls.source + ": " + line + ".    " + err)
}

func (ls *LexState) incLine() {
	old := ls.current
	ls.next() /* skip '\n' or '\r' */
	if isNewLine(ls.current) && ls.current != old {
		ls.next() /* skip '\n\r' or '\r\n' */
	}
	ls.lineNumber++
}

func (ls *LexState) readNumber() (token.Type, *token.SemInfo) {
	hasDot := false
	isHex := false
	sem := &token.SemInfo{}
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
		return token.Float, sem
	}
	i, err := strconv.ParseInt(sem.S, 0, 64)
	if err != nil {
		ls.lexErr(err.Error())
	}
	sem.I = i
	return token.Integer, sem
}

func (ls *LexState) readIdent() (token.Type, *token.SemInfo) {
	sem := &token.SemInfo{}
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
		if strings.Count(sem.S, "::") == 2 && strings.Count(sem.S, ":") == 4 {
			sem.S = sem.S[strings.Index(sem.S, "::")+2:]
		}
		if strings.Count(sem.S, "::") != 1 || strings.Count(sem.S, ":") != 2 {
			ls.lexErr("namespace qualifier::is illegal")
		}
	}

	for i := token.DummyKeywordBegin + 1; i < token.DummyKeywordEnd; i++ {
		if token.Value(i) == sem.S {
			return i, nil
		}
	}
	for i := token.DummyTypeBegin + 1; i < token.DummyTypeEnd; i++ {
		if token.Value(i) == sem.S {
			return i, nil
		}
	}

	return token.Name, sem
}

func (ls *LexState) readSharp() (token.Type, *token.SemInfo) {
	ls.next()
	for isLetter(ls.current) {
		ls.tokenBuff.WriteByte(ls.current)
		ls.next()
	}
	if ls.tokenBuff.String() != "include" {
		ls.lexErr("not #include")
	}

	return token.Include, nil
}

func (ls *LexState) readString() (token.Type, *token.SemInfo) {
	sem := &token.SemInfo{}
	ls.next()
	for {
		if ls.current == token.EOF {
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

	return token.String, sem
}

func (ls *LexState) readLongComment() {
	for {
		switch ls.current {
		case token.EOF:
			ls.lexErr("respect */")
			return
		case '\n', '\r':
			ls.incLine()
		case '*':
			ls.next()
			if ls.current == token.EOF {
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
		ls.current = token.EOF
	}
}

func (ls *LexState) llexDefault() (token.Type, *token.SemInfo) {
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
func (ls *LexState) lLex() (token.Type, *token.SemInfo) {
	for {
		ls.tokenBuff.Reset()
		switch ls.current {
		case token.EOF:
			return token.Eof, nil
		case ' ', '\t', '\f', '\v':
			ls.next()
		case '\n', '\r':
			ls.incLine()
		case '/': // Comment processing
			ls.next()
			if ls.current == '/' {
				for !isNewLine(ls.current) && ls.current != token.EOF {
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
			return token.BraceLeft, nil
		case '}':
			ls.next()
			return token.BraceRight, nil
		case ';':
			ls.next()
			return token.Semi, nil
		case '=':
			ls.next()
			return token.Eq, nil
		case '<':
			ls.next()
			return token.Shl, nil
		case '>':
			ls.next()
			return token.Shr, nil
		case ',':
			ls.next()
			return token.Comma, nil
		case '(':
			ls.next()
			return token.Ptl, nil
		case ')':
			ls.next()
			return token.Ptr, nil
		case '[':
			ls.next()
			return token.SquareLeft, nil
		case ']':
			ls.next()
			return token.SquarerRight, nil
		case '"':
			return ls.readString()
		case '#':
			return ls.readSharp()
		default:
			return ls.llexDefault()

		}
	}
}

// NextToken return token after lexical analysis.
func (ls *LexState) NextToken() *token.Token {
	tk := &token.Token{}
	tk.T, tk.S = ls.lLex()
	tk.Line = ls.lineNumber
	return tk
}

// NewLexState to update LexState struct.
func NewLexState(source string, buff []byte) *LexState {
	return &LexState{
		current:    ' ',
		lineNumber: 1,
		source:     source,
		buff:       bytes.NewBuffer(buff),
	}
}
