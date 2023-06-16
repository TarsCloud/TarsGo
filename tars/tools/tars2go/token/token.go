package token

// Type is a byte type.
type Type byte

// SemInfo is struct.
type SemInfo struct {
	I int64
	F float64
	S string
}

// Token record token information.
type Token struct {
	T    Type
	S    *SemInfo
	Line int
}

const EOF = 0

const (
	Eof          Type = iota
	BraceLeft         //({)
	BraceRight        //}
	Semi              //;
	Eq                //=
	Shl               //<
	Shr               //>
	Comma             //,
	Ptl               //(
	Ptr               //)
	SquareLeft        //[
	SquarerRight      //]
	Include           //#include

	// DummyKeywordBegin keyword
	DummyKeywordBegin
	Module
	Enum
	Struct
	Interface
	Require
	Optional
	Const
	Unsigned
	Void
	Out
	Key
	True
	False
	DummyKeywordEnd

	// DummyTypeBegin type
	DummyTypeBegin
	TInt
	TBool
	TShort
	TByte
	TLong
	TFloat
	TDouble
	TString
	TVector
	TMap
	TArray
	DummyTypeEnd

	Name // variable name
	// String value
	String
	Integer
	Float
)

// tokenMap record token value.
var tokenMap = [...]string{
	Eof: "<eos>",

	BraceLeft:    "{",
	BraceRight:   "}",
	Semi:         ";",
	Eq:           "=",
	Shl:          "<",
	Shr:          ">",
	Comma:        ",",
	Ptl:          "(",
	Ptr:          ")",
	SquareLeft:   "[",
	SquarerRight: "]",
	Include:      "#include",

	// keyword
	Module:    "module",
	Enum:      "enum",
	Struct:    "struct",
	Interface: "interface",
	Require:   "require",
	Optional:  "optional",
	Const:     "const",
	Unsigned:  "unsigned",
	Void:      "void",
	Out:       "out",
	Key:       "key",
	True:      "true",
	False:     "false",

	// type
	TInt:    "int",
	TBool:   "bool",
	TShort:  "short",
	TByte:   "byte",
	TLong:   "long",
	TFloat:  "float",
	TDouble: "double",
	TString: "string",
	TVector: "vector",
	TMap:    "map",
	TArray:  "array",

	Name: "<name>",
	// value
	String:  "<string>",
	Integer: "<INTEGER>",
	Float:   "<FLOAT>",
}

func Value(typ Type) string {
	return tokenMap[typ]
}

func IsType(typ Type) bool {
	return typ > DummyTypeBegin && typ < DummyTypeEnd
}

func IsNumberType(typ Type) bool {
	switch typ {
	case TInt, TBool, TShort, TByte, TLong, TFloat, TDouble:
		return true
	default:
		return false
	}
}
