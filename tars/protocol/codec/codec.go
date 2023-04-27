// Package codec implement
// 支持tars2go的底层库，用于基础类型的序列化
// 高级类型的序列化，由代码生成器，转换为基础类型的序列化

package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"
)

// jce type
const (
	BYTE byte = iota
	SHORT
	INT
	LONG
	FLOAT
	DOUBLE
	STRING1
	STRING4
	MAP
	LIST
	StructBegin
	StructEnd
	ZeroTag
	SimpleList
)

var typeToStr = []string{
	"Byte",
	"Short",
	"Int",
	"Long",
	"Float",
	"Double",
	"String1",
	"String4",
	"Map",
	"List",
	"StructBegin",
	"StructEnd",
	"ZeroTag",
	"SimpleList",
}

func getTypeStr(t int) string {
	if t < len(typeToStr) {
		return typeToStr[t]
	}
	return "invalidType"
}

// Buffer is wrapper of bytes.Buffer
type Buffer struct {
	buf *bytes.Buffer
}

// Reader is wrapper of bytes.Reader
type Reader struct {
	ref []byte
	buf *bytes.Reader
}

//go:nosplit
func bWriteU8(w *bytes.Buffer, data uint8) error {
	err := w.WriteByte(data)
	return err
}

//go:nosplit
func bWriteU16(w *bytes.Buffer, data uint16) error {
	var (
		b  [2]byte
		bs []byte
	)
	bs = b[:]
	binary.BigEndian.PutUint16(bs, data)
	_, err := w.Write(bs)
	return err
}

//go:nosplit
func bWriteU32(w *bytes.Buffer, data uint32) error {
	var (
		b  [4]byte
		bs []byte
	)
	bs = b[:]
	binary.BigEndian.PutUint32(bs, data)
	_, err := w.Write(bs)
	return err
}

//go:nosplit
func bWriteU64(w *bytes.Buffer, data uint64) error {
	var (
		b  [8]byte
		bs []byte
	)
	bs = b[:]
	binary.BigEndian.PutUint64(bs, data)
	_, err := w.Write(bs)
	return err
}

//go:nosplit
func bReadU8(r *bytes.Reader, data *uint8) error {
	var err error
	*data, err = r.ReadByte()
	return err
}

//go:nosplit
func bReadU16(r *bytes.Reader, data *uint16) error {
	var (
		b  [2]byte
		bs []byte
	)
	bs = b[:]
	_, err := r.Read(bs)
	*data = binary.BigEndian.Uint16(bs)
	return err
}

//go:nosplit
func bReadU32(r *bytes.Reader, data *uint32) error {
	var (
		b  [4]byte
		bs []byte
	)
	bs = b[:]
	_, err := r.Read(bs)
	*data = binary.BigEndian.Uint32(bs)
	return err
}

//go:nosplit
func bReadU64(r *bytes.Reader, data *uint64) error {
	var (
		b  [8]byte
		bs []byte
	)
	bs = b[:]
	_, err := r.Read(bs)
	*data = binary.BigEndian.Uint64(bs)
	return err
}

//go:nosplit
func (b *Buffer) WriteHead(ty byte, tag byte) error {
	if tag < 15 {
		data := (tag << 4) | ty
		return b.buf.WriteByte(data)
	} else {
		data := (15 << 4) | ty
		if err := b.buf.WriteByte(data); err != nil {
			return err
		}
		return b.buf.WriteByte(tag)
	}
}

// Reset clean the buffer.
func (b *Buffer) Reset() {
	b.buf.Reset()
}

// WriteSliceUint8 write []uint8 to the buffer.
func (b *Buffer) WriteSliceUint8(data []uint8) error {
	_, err := b.buf.Write(data)
	return err
}

// WriteSliceInt8 write []int8 to the buffer.
func (b *Buffer) WriteSliceInt8(data []int8) error {
	_, err := b.buf.Write(*(*[]uint8)(unsafe.Pointer(&data)))
	return err
}

// WriteBytes write []byte to the buffer
func (b *Buffer) WriteBytes(data []byte) error {
	_, err := b.buf.Write(data)
	return err
}

// WriteInt8 write int8 with the tag.
func (b *Buffer) WriteInt8(data int8, tag byte) error {
	var err error
	if data == 0 {
		if err = b.WriteHead(ZeroTag, tag); err != nil {
			return err
		}
	} else {
		if err = b.WriteHead(BYTE, tag); err != nil {
			return err
		}

		if err = b.buf.WriteByte(byte(data)); err != nil {
			return err
		}
	}
	return nil
}

// WriteUint8 write uint8 with the tag
func (b *Buffer) WriteUint8(data uint8, tag byte) error {
	return b.WriteInt16(int16(data), tag)
}

// WriteBool write bool with the tag.
func (b *Buffer) WriteBool(data bool, tag byte) error {
	tmp := int8(0)
	if data {
		tmp = 1
	}
	return b.WriteInt8(tmp, tag)
}

// WriteInt16 write the int16 with the tag.
func (b *Buffer) WriteInt16(data int16, tag byte) error {
	var err error
	if data >= math.MinInt8 && data <= math.MaxInt8 {
		if err = b.WriteInt8(int8(data), tag); err != nil {
			return err
		}
	} else {
		if err = b.WriteHead(SHORT, tag); err != nil {
			return err
		}

		if err = bWriteU16(b.buf, uint16(data)); err != nil {
			return err
		}
	}
	return nil
}

// WriteUint16 write uint16 with the tag.
func (b *Buffer) WriteUint16(data uint16, tag byte) error {
	return b.WriteInt32(int32(data), tag)
}

// WriteInt32 write int32 with the tag.
func (b *Buffer) WriteInt32(data int32, tag byte) error {
	var err error
	if data >= math.MinInt16 && data <= math.MaxInt16 {
		if err = b.WriteInt16(int16(data), tag); err != nil {
			return err
		}
	} else {
		if err = b.WriteHead(INT, tag); err != nil {
			return err
		}

		if err = bWriteU32(b.buf, uint32(data)); err != nil {
			return err
		}
	}
	return nil
}

// WriteUint32 write uint32 data with the tag.
func (b *Buffer) WriteUint32(data uint32, tag byte) error {
	return b.WriteInt64(int64(data), tag)
}

// WriteInt64 write int64 with the tag.
func (b *Buffer) WriteInt64(data int64, tag byte) error {
	var err error
	if data >= math.MinInt32 && data <= math.MaxInt32 {
		if err = b.WriteInt32(int32(data), tag); err != nil {
			return err
		}
	} else {
		if err = b.WriteHead(LONG, tag); err != nil {
			return err
		}

		if err = bWriteU64(b.buf, uint64(data)); err != nil {
			return err
		}
	}
	return nil
}

// WriteFloat32 writes float32 with the tag.
func (b *Buffer) WriteFloat32(data float32, tag byte) error {
	var err error
	if err = b.WriteHead(FLOAT, tag); err != nil {
		return err
	}

	err = bWriteU32(b.buf, math.Float32bits(data))
	return err
}

// WriteFloat64 writes float64 with the tag.
func (b *Buffer) WriteFloat64(data float64, tag byte) error {
	var err error
	if err = b.WriteHead(DOUBLE, tag); err != nil {
		return err
	}

	err = bWriteU64(b.buf, math.Float64bits(data))
	return err
}

// WriteString writes string data with the tag.
func (b *Buffer) WriteString(data string, tag byte) error {
	var err error
	if len(data) > 255 {
		if err = b.WriteHead(STRING4, tag); err != nil {
			return err
		}

		if err = bWriteU32(b.buf, uint32(len(data))); err != nil {
			return err
		}
	} else {
		if err = b.WriteHead(STRING1, tag); err != nil {
			return err
		}

		if err = bWriteU8(b.buf, byte(len(data))); err != nil {
			return err
		}
	}

	if _, err = b.buf.WriteString(data); err != nil {
		return err
	}
	return nil
}

// ToBytes make the buffer to []byte
func (b *Buffer) ToBytes() []byte {
	return b.buf.Bytes()
}

func (b *Buffer) Len() int {
	return b.buf.Len()
}

// Grow grows the size of the buffer.
func (b *Buffer) Grow(size int) {
	b.buf.Grow(size)
}

// Reset clean the Reader.
func (b *Reader) Reset(data []byte) {
	b.buf.Reset(data)
	b.ref = data
}

//go:nosplit
func (b *Reader) readHead() (ty, tag byte, err error) {
	data, err := b.buf.ReadByte()
	if err != nil {
		return
	}
	ty = data & 0x0f
	tag = (data & 0xf0) >> 4
	if tag == 15 {
		data, err = b.buf.ReadByte()
		if err != nil {
			return
		}
		tag = data
	}
	return
}

// unreadHead 回退一个head byte， curTag为当前读到的tag信息，当tag超过4位时则回退两个head byte
// unreadHead put back the current head byte.
func (b *Reader) unreadHead(curTag byte) {
	_ = b.buf.UnreadByte()
	if curTag >= 15 {
		_ = b.buf.UnreadByte()
	}
}

// Next return the []byte of next n .
//
//go:nosplit
func (b *Reader) Next(n int) []byte {
	if n <= 0 {
		return []byte{}
	}
	beg := len(b.ref) - b.buf.Len()
	_, _ = b.buf.Seek(int64(n), io.SeekCurrent)
	end := len(b.ref) - b.buf.Len()
	return b.ref[beg:end]
}

// Skip the next n byte.
//
//go:nosplit
func (b *Reader) Skip(n int) {
	if n <= 0 {
		return
	}
	_, _ = b.buf.Seek(int64(n), io.SeekCurrent)
}

func (b *Reader) skipFieldMap() error {
	var length int32
	err := b.ReadInt32(&length, 0, true)
	if err != nil {
		return err
	}

	for i := int32(0); i < length*2; i++ {
		tyCur, _, err := b.readHead()
		if err != nil {
			return err
		}
		_ = b.skipField(tyCur)
	}
	return nil
}
func (b *Reader) skipFieldList() error {
	var length int32
	err := b.ReadInt32(&length, 0, true)
	if err != nil {
		return err
	}
	for i := int32(0); i < length; i++ {
		tyCur, _, err := b.readHead()
		if err != nil {
			return err
		}
		_ = b.skipField(tyCur)
	}
	return nil
}
func (b *Reader) skipFieldSimpleList() error {
	tyCur, _, err := b.readHead()
	if tyCur != BYTE {
		return fmt.Errorf("simple list need byte head. but get %d", tyCur)
	}
	if err != nil {
		return err
	}
	var length int32
	err = b.ReadInt32(&length, 0, true)
	if err != nil {
		return err
	}

	b.Skip(int(length))
	return nil
}

func (b *Reader) skipField(ty byte) error {
	switch ty {
	case BYTE:
		b.Skip(1)
	case SHORT:
		b.Skip(2)
	case INT:
		b.Skip(4)
	case LONG:
		b.Skip(8)
	case FLOAT:
		b.Skip(4)
	case DOUBLE:
		b.Skip(8)
	case STRING1:
		data, err := b.buf.ReadByte()
		if err != nil {
			return err
		}
		l := int(data)
		b.Skip(l)
	case STRING4:
		var l uint32
		err := bReadU32(b.buf, &l)
		if err != nil {
			return err
		}
		b.Skip(int(l))
	case MAP:
		err := b.skipFieldMap()
		if err != nil {
			return err
		}
	case LIST:
		err := b.skipFieldList()
		if err != nil {
			return err
		}
	case SimpleList:
		err := b.skipFieldSimpleList()
		if err != nil {
			return err
		}
	case StructBegin:
		err := b.SkipToStructEnd()
		if err != nil {
			return err
		}
	case StructEnd:
	case ZeroTag:
	default:
		return fmt.Errorf("invalid type")
	}
	return nil
}

// SkipToStructEnd for skip to the StructEnd tag.
func (b *Reader) SkipToStructEnd() error {
	for {
		ty, _, err := b.readHead()
		if err != nil {
			return err
		}

		err = b.skipField(ty)
		if err != nil {
			return err
		}
		if ty == StructEnd {
			break
		}
	}
	return nil
}

// SkipToNoCheck for skip to the none StructEnd tag.
func (b *Reader) SkipToNoCheck(tag byte, require bool) (bool, byte, error) {
	for {
		tyCur, tagCur, err := b.readHead()
		if err != nil {
			if require {
				return false, tyCur, fmt.Errorf("can not find Tag %d. But require. %s", tag, err.Error())
			}
			return false, tyCur, nil
		}
		if tyCur == StructEnd || tagCur > tag {
			if require {
				return false, tyCur, fmt.Errorf("can not find Tag %d. But require. tagCur: %d, tyCur: %d",
					tag, tagCur, tyCur)
			}
			// 多读了一个head, 退回去.
			b.unreadHead(tagCur)
			return false, tyCur, nil
		}
		if tagCur == tag {
			return true, tyCur, nil
		}

		// tagCur < tag
		if err = b.skipField(tyCur); err != nil {
			return false, tyCur, err
		}
	}
}

// SkipTo skip to the given tag.
func (b *Reader) SkipTo(ty, tag byte, require bool) (bool, error) {
	have, tyCur, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return false, err
	}
	if have && ty != tyCur {
		return false, fmt.Errorf("type not match, need %d, bug %d", ty, tyCur)
	}
	return have, nil
}

// ReadSliceInt8 reads []int8 for the given length and the require or optional sign.
func (b *Reader) ReadSliceInt8(data *[]int8, len int32, require bool) error {
	if len <= 0 {
		return nil
	}

	*data = make([]int8, len)
	_, err := b.buf.Read(*(*[]uint8)(unsafe.Pointer(data)))
	if err != nil {
		err = fmt.Errorf("read []int8 error:%v", err)
	}
	return err
}

// ReadSliceUint8 reads []uint8 force the given length and the require or optional sign.
func (b *Reader) ReadSliceUint8(data *[]uint8, len int32, require bool) error {
	if len <= 0 {
		return nil
	}

	*data = make([]uint8, len)
	_, err := b.buf.Read(*data)
	if err != nil {
		err = fmt.Errorf("read []uint8 error:%v", err)
	}
	return err
}

// ReadBytes reads []byte for the given length and the require or optional sign.
func (b *Reader) ReadBytes(data *[]byte, len int32, require bool) error {
	*data = make([]byte, len)
	_, err := b.buf.Read(*data)
	return err
}

// ReadInt8 reads the int8 data for the tag and the require or optional sign.
func (b *Reader) ReadInt8(data *int8, tag byte, require bool) error {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return err
	}
	if !have {
		return nil
	}
	switch ty {
	case ZeroTag:
		*data = 0
	case BYTE:
		var tmp uint8
		err = bReadU8(b.buf, &tmp)
		*data = int8(tmp)
	default:
		return fmt.Errorf("read 'int8' type mismatch, tag:%d, get type:%s", tag, getTypeStr(int(ty)))
	}
	if err != nil {
		err = fmt.Errorf("read 'int8' tag:%d error:%v", tag, err)
	}
	return err
}

// ReadUint8 reads the uint8 for the tag and the require or optional sign.
func (b *Reader) ReadUint8(data *uint8, tag byte, require bool) error {
	n := int16(*data)
	err := b.ReadInt16(&n, tag, require)
	*data = uint8(n)
	return err
}

// ReadBool reads the bool value for the tag and the require or optional sign.
func (b *Reader) ReadBool(data *bool, tag byte, require bool) error {
	var tmp int8
	if *data {
		tmp = 1
	}
	err := b.ReadInt8(&tmp, tag, require)
	if err != nil {
		return err
	}
	if tmp == 0 {
		*data = false
	} else {
		*data = true
	}
	return nil
}

// ReadInt16 reads the int16 value for the tag and the require or optional sign.
func (b *Reader) ReadInt16(data *int16, tag byte, require bool) error {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return err
	}
	if !have {
		return nil
	}
	switch ty {
	case ZeroTag:
		*data = 0
	case BYTE:
		var tmp uint8
		err = bReadU8(b.buf, &tmp)
		*data = int16(int8(tmp))
	case SHORT:
		var tmp uint16
		err = bReadU16(b.buf, &tmp)
		*data = int16(tmp)
	default:
		return fmt.Errorf("read 'int16' type mismatch, tag:%d, get type:%s", tag, getTypeStr(int(ty)))
	}
	if err != nil {
		err = fmt.Errorf("read 'int16' tag:%d error:%v", tag, err)
	}
	return err
}

// ReadUint16 reads the uint16 value for the tag and the require or optional sign.
func (b *Reader) ReadUint16(data *uint16, tag byte, require bool) error {
	n := int32(*data)
	err := b.ReadInt32(&n, tag, require)
	*data = uint16(n)
	return err
}

// ReadInt32 reads the int32 value for the tag and the require or optional sign.
func (b *Reader) ReadInt32(data *int32, tag byte, require bool) error {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return err
	}
	if !have {
		return nil
	}
	switch ty {
	case ZeroTag:
		*data = 0
	case BYTE:
		var tmp uint8
		err = bReadU8(b.buf, &tmp)
		*data = int32(int8(tmp))
	case SHORT:
		var tmp uint16
		err = bReadU16(b.buf, &tmp)
		*data = int32(int16(tmp))
	case INT:
		var tmp uint32
		err = bReadU32(b.buf, &tmp)
		*data = int32(tmp)
	default:
		return fmt.Errorf("read 'int32' type mismatch, tag:%d, get type:%s", tag, getTypeStr(int(ty)))
	}
	if err != nil {
		err = fmt.Errorf("read 'int32' tag:%d error:%v", tag, err)
	}
	return err
}

// ReadUint32 reads the uint32 value for the tag and the require or optional sign.
func (b *Reader) ReadUint32(data *uint32, tag byte, require bool) error {
	n := int64(*data)
	err := b.ReadInt64(&n, tag, require)
	*data = uint32(n)
	return err
}

// ReadInt64 reads the int64 value for the tag and the require or optional sign.
func (b *Reader) ReadInt64(data *int64, tag byte, require bool) error {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return err
	}
	if !have {
		return nil
	}
	switch ty {
	case ZeroTag:
		*data = 0
	case BYTE:
		var tmp uint8
		err = bReadU8(b.buf, &tmp)
		*data = int64(int8(tmp))
	case SHORT:
		var tmp uint16
		err = bReadU16(b.buf, &tmp)
		*data = int64(int16(tmp))
	case INT:
		var tmp uint32
		err = bReadU32(b.buf, &tmp)
		*data = int64(int32(tmp))
	case LONG:
		var tmp uint64
		err = bReadU64(b.buf, &tmp)
		*data = int64(tmp)
	default:
		return fmt.Errorf("read 'int64' type mismatch, tag:%d, get type:%s", tag, getTypeStr(int(ty)))
	}
	if err != nil {
		err = fmt.Errorf("read 'int64' tag:%d error:%v", tag, err)
	}

	return err
}

// ReadFloat32 reads the float32 value for the tag and the require or optional sign.
func (b *Reader) ReadFloat32(data *float32, tag byte, require bool) error {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return err
	}
	if !have {
		return nil
	}

	switch ty {
	case ZeroTag:
		*data = 0
	case FLOAT:
		var tmp uint32
		err = bReadU32(b.buf, &tmp)
		*data = math.Float32frombits(tmp)
	default:
		return fmt.Errorf("read 'float' type mismatch, tag:%d, get type:%s", tag, getTypeStr(int(ty)))
	}

	if err != nil {
		err = fmt.Errorf("read 'float32' tag:%d error:%v", tag, err)
	}
	return err
}

// ReadFloat64 reads the float64 value for the tag and the require or optional sign.
func (b *Reader) ReadFloat64(data *float64, tag byte, require bool) error {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return err
	}
	if !have {
		return nil
	}

	switch ty {
	case ZeroTag:
		*data = 0
	case FLOAT:
		var tmp uint32
		err = bReadU32(b.buf, &tmp)
		*data = float64(math.Float32frombits(tmp))
	case DOUBLE:
		var tmp uint64
		err = bReadU64(b.buf, &tmp)
		*data = math.Float64frombits(tmp)
	default:
		return fmt.Errorf("read 'double' type mismatch, tag:%d, get type:%s", tag, getTypeStr(int(ty)))
	}

	if err != nil {
		err = fmt.Errorf("read 'float64' tag:%d error:%v", tag, err)
	}
	return err
}

// ReadString reads the string value for the tag and the require or optional sign.
func (b *Reader) ReadString(data *string, tag byte, require bool) error {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return err
	}
	if !have {
		return nil
	}

	if ty == STRING4 {
		var length uint32
		err = bReadU32(b.buf, &length)
		if err != nil {
			return fmt.Errorf("read string4 tag:%d error:%v", tag, err)
		}
		buff := b.Next(int(length))
		*data = string(buff)
	} else if ty == STRING1 {
		var length uint8
		err = bReadU8(b.buf, &length)
		if err != nil {
			return fmt.Errorf("read string1 tag:%d error:%v", tag, err)
		}
		buff := b.Next(int(length))
		*data = string(buff)
	} else {
		return fmt.Errorf("need string, tag:%d, but type is %s", tag, getTypeStr(int(ty)))
	}
	return nil
}

// ToString make the reader to string
func (b *Reader) ToString() string {
	return string(b.ref[:])
}

// ToBytes make the reader to string
func (b *Reader) ToBytes() []byte {
	return b.ref
}

func (b *Reader) Len() int {
	return len(b.ref)
}

// NewReader returns *Reader
func NewReader(data []byte) *Reader {
	return &Reader{buf: bytes.NewReader(data), ref: data}
}

// NewBuffer returns *Buffer
func NewBuffer(args ...*bytes.Buffer) *Buffer {
	buf := &bytes.Buffer{}
	if len(args) > 0 {
		buf = args[0]
	}
	return &Buffer{buf: buf}
}

// FromInt8 NewReader(FromInt8(vec))
func FromInt8(vec []int8) []byte {
	return *(*[]byte)(unsafe.Pointer(&vec))
}
