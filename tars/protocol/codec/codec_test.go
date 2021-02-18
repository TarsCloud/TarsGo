package codec

import (
	"math"
	"math/rand"
	"reflect"
	"testing"
)

func r(b *Buffer) *Reader {
	buf := b.ToBytes()
	return NewReader(buf)
}

// TestUint8 tests the read and write of the uint8 type.
func TestUint8(t *testing.T) {
	for tag := 0; tag < 250; tag++ {
		for i := 0; i <= math.MaxUint8; i++ {
			b := NewBuffer()
			err := b.Write_uint8(uint8(i), byte(tag))
			if err != nil {
				t.Error(err)
			}
			var data uint8
			err = r(b).Read_uint8(&data, byte(tag), true)
			if err != nil {
				t.Error(err)
			}
			if data != uint8(i) {
				t.Error("no eq.")
			}
		}
	}
}

// TestInt8 tests the read and write of the int8 type.
func TestInt8(t *testing.T) {
	for tag := 0; tag < 250; tag++ {
		for i := math.MinInt8; i <= math.MaxInt8; i++ {
			b := NewBuffer()
			err := b.Write_int8(int8(i), byte(tag))
			if err != nil {
				t.Error(err)
			}
			var data int8
			err = r(b).Read_int8(&data, byte(tag), true)
			if err != nil {
				t.Error(err)
			}
			if data != int8(i) {
				t.Error("no eq.")
			}
		}
	}
}

// TestUint16 tests the read and write of the int16 type.
func TestUint16(t *testing.T) {
	for tag := 0; tag < 250; tag += 10 {
		for i := 0; i < math.MaxUint16; i++ {
			b := NewBuffer()
			err := b.Write_uint16(uint16(i), byte(tag))
			if err != nil {
				t.Error(err)
			}
			var data uint16
			err = r(b).Read_uint16(&data, byte(tag), true)
			if err != nil {
				t.Error(err)
			}
			if data != uint16(i) {
				t.Error("no eq.")
			}
		}
	}
}

// TestInt16 tests the read and write of the int16  type.
func TestInt16(t *testing.T) {
	for tag := 0; tag < 250; tag += 10 {
		for i := math.MinInt16; i <= math.MaxInt16; i++ {
			b := NewBuffer()
			err := b.Write_int16(int16(i), byte(tag))
			if err != nil {
				t.Error(err)
			}
			var data int16
			err = r(b).Read_int16(&data, byte(tag), true)
			if err != nil {
				t.Error(err)
			}
			if data != int16(i) {
				t.Error("no eq.", i)
			}
		}
	}
}

// TestInt16_2 tests the read and write of the int16  type.
func TestInt16_2(t *testing.T) {
	b := NewBuffer()
	err := b.Write_int16(int16(-1), byte(0))
	if err != nil {
		t.Error(err)
	}
	var data int16
	err = r(b).Read_int16(&data, byte(0), true)
	if err != nil {
		t.Error(err)
	}
	if data != int16(-1) {
		t.Error("no eq.", data)
	}
}

// TestInt32  tests the read and write of the int32  type.
func TestInt32(t *testing.T) {
	b := NewBuffer()
	err := b.Write_int32(int32(-1), byte(10))
	if err != nil {
		t.Error(err)
	}
	var data int32
	err = r(b).Read_int32(&data, byte(10), true)
	if err != nil {
		t.Error(err)
	}
	if data != -1 {
		t.Error("no eq.")
	}
}

// TestInt32_2  tests the read and write of the int32  type.
func TestInt32_2(t *testing.T) {
	b := NewBuffer()
	err := b.Write_int32(math.MinInt32, byte(10))
	if err != nil {
		t.Error(err)
	}
	var data int32
	err = r(b).Read_int32(&data, byte(10), true)
	if err != nil {
		t.Error(err)
	}
	if data != math.MinInt32 {
		t.Error("no eq.")
	}
}

// TestUint32  tests the read and write of the uint32  type.
func TestUint32(t *testing.T) {
	b := NewBuffer()
	err := b.Write_uint32(uint32(0xffffffff), byte(10))
	if err != nil {
		t.Error(err)
	}
	var data uint32
	err = r(b).Read_uint32(&data, byte(10), true)
	if err != nil {
		t.Error(err)
	}
	if data != 0xffffffff {
		t.Error("no eq.")
	}
}

// TestInt64 tests the read and write of the int64  type.
func TestInt64(t *testing.T) {
	b := NewBuffer()
	err := b.Write_int64(math.MinInt64, byte(10))
	if err != nil {
		t.Error(err)
	}
	var data int64
	err = r(b).Read_int64(&data, byte(10), true)
	if err != nil {
		t.Error(err)
	}
	if data != math.MinInt64 {
		t.Error("no eq.")
	}
}

// TestSkipString tests skip the string.
func TestSkipString(t *testing.T) {
	b := NewBuffer()
	for i := 0; i < 200; i++ {
		bs := make([]byte, 200+i)
		err := b.Write_string(string(bs), byte(i))
		if err != nil {
			t.Error(err)
		}
	}

	var data string
	err := r(b).Read_string(&data, byte(190), true)
	if err != nil {
		t.Error(err)
	}
	bs := make([]byte, 200+190)
	if data != string(bs) {
		t.Error("no eq.")
	}
}

// TestSkipStruct tests skip struct.
func TestSkipStruct(t *testing.T) {
	b := NewBuffer()

	err := b.WriteHead(STRUCT_BEGIN, 1)
	if err != nil {
		t.Error(err)
	}

	err = b.WriteHead(STRUCT_END, 0)
	if err != nil {
		t.Error(err)
	}

	rd := r(b)

	err, have := rd.SkipTo(STRUCT_BEGIN, 1, true)
	if err != nil || have == false {
		t.Error(err)
	}
	err = rd.SkipToStructEnd()
	if err != nil || have == false {
		t.Error(err)
	}
}

// TestSkipStruct2 tests skip struct.
func TestSkipStruct2(t *testing.T) {
	b := NewBuffer()

	err := b.WriteHead(STRUCT_BEGIN, 1)
	if err != nil {
		t.Error(err)
	}
	err = b.WriteHead(STRUCT_BEGIN, 1)
	if err != nil {
		t.Error(err)
	}

	err = b.WriteHead(STRUCT_END, 0)
	if err != nil {
		t.Error(err)
	}
	err = b.WriteHead(STRUCT_END, 0)
	if err != nil {
		t.Error(err)
	}
	err = b.Write_int64(math.MinInt64, byte(10))
	if err != nil {
		t.Error(err)
	}

	rb := r(b)

	err, have := rb.SkipTo(STRUCT_BEGIN, 1, true)
	if err != nil || !have {
		t.Error(err)
	}
	err = rb.SkipToStructEnd()
	if err != nil {
		t.Error(err)
	}
	var data int64
	err = rb.Read_int64(&data, byte(10), true)
	if err != nil {
		t.Error(err)
	}
	if data != math.MinInt64 {
		t.Error("no eq.")
	}
}

// BenchmarkUint32 benchmarks the write and read the uint32 type.
func BenchmarkUint32(t *testing.B) {
	b := NewBuffer()

	for i := 0; i < 200; i++ {
		err := b.Write_uint32(uint32(0xffffffff), byte(i))
		if err != nil {
			t.Error(err)
		}
	}

	rb := r(b)

	for i := 0; i < 200; i++ {
		var data uint32
		err := rb.Read_uint32(&data, byte(i), true)
		if err != nil {
			t.Error(err)
		}
		if data != 0xffffffff {
			t.Error("no eq.")
		}
	}
}

// BenchmarkString benchmark the read and write the string.
func BenchmarkString(t *testing.B) {
	b := NewBuffer()

	for i := 0; i < 200; i++ {
		err := b.Write_string("hahahahahahahahahahahahahahahahahahahaha", byte(i))
		if err != nil {
			t.Error(err)
		}
	}

	rb := r(b)

	for i := 0; i < 200; i++ {
		var data string
		err := rb.Read_string(&data, byte(i), true)
		if err != nil {
			t.Error(err)
		}
		if data != "hahahahahahahahahahahahahahahahahahahaha" {
			t.Error("no eq.")
		}
	}
}

func TestBuffer_Reset(t *testing.T) {
	got := NewBuffer()
	err := got.Write_bytes([]byte("test"))
	if err != nil {
		t.Errorf("Write bytes failed")
	}
	got.Reset()

	want := NewBuffer()

	if want.buf.String() != got.buf.String() || want.buf.Len() != got.buf.Len() {
		t.Errorf("Test Reset failed. want:'%v', len=%v got:'%v', len=%v\n",
			want.buf.String(), want.buf.Len(), got.buf.String(), got.buf.Len())
	}
}

func TestBuffer_slice_uint8(t *testing.T) {
	var want = []uint8{1, 2, 3, 4, 5}
	var got []uint8
	buf := NewBuffer()
	err := buf.Write_slice_uint8(want)
	if err != nil {
		t.Errorf("Test Write_slice_uint8 failed. err:%s\n", err)
	}
	reader := r(buf)
	err = reader.Read_slice_uint8(&got, int32(len(want)), true)
	if err != nil {
		t.Errorf("Test Read_slice_uint8 failed. err:%s\n", err)
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Test Write_slice_uint8 failed. want:%v, got:%v\n", want, got)
	}
}

func TestBuffer_slice_int8(t *testing.T) {
	var want = []int8{1, 2, 3, 4, 5}
	var got []int8
	buf := NewBuffer()
	err := buf.Write_slice_int8(want)
	if err != nil {
		t.Errorf("Test Write_slice_int8 failed. err:%s\n", err)
	}
	reader := r(buf)
	err = reader.Read_slice_int8(&got, int32(len(want)), true)
	if err != nil {
		t.Errorf("Test Read_slice_int8 failed. err:%s\n", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Test Write_slice_int8 failed. want:%v, got:%v\n", want, got)
	}
}

func TestBuffer_bytes(t *testing.T) {
	var wants = [][]byte{
		{1, 3, 4, 5, 6, 100, 250},
		[]byte("hello world"),
	}
	var got []byte
	for _, want := range wants {
		buf := NewBuffer()
		err := buf.Write_bytes(want)
		if err != nil {
			t.Errorf("Test Write_bytes failed. err:%s\n", err)
		}

		reader := r(buf)
		err = reader.Read_bytes(&got, int32(len(want)), true)
		if err != nil {
			t.Errorf("Test Read_bytes failed. err:%s\n", err)
		}

		if !reflect.DeepEqual(want, got) {
			t.Errorf("Test Write_bytes failed. want:%v, got:%v\n", want, got)
		}
	}
}

func TestBuffer_bool(t *testing.T) {
	var got bool
	wants := []bool{true, false}
	for _, want := range wants {
		buf := NewBuffer()
		err := buf.Write_bool(want, 10)
		if err != nil {
			t.Errorf("Test Write_bool failed. err:%s\n", err)
		}

		reader := r(buf)
		err = reader.Read_bool(&got, 10, true)
		if err != nil {
			t.Errorf("Test Read_bool failed. err:%s\n", err)
		}
		if got != want {
			t.Errorf("Test Write_bool failed, want:%v, got:%v\n", want, got)
		}
	}
}

func TestBuffer_float32(t *testing.T) {
	got := float32(0)
	for i := 0; i < 500; i++ {
		writer := NewBuffer()
		want := rand.Float32()

		err := writer.Write_float32(want, 3)
		if err != nil {
			t.Errorf("Test Write_float32 failed. err:%s\n", err)
		}

		reader := r(writer)
		err = reader.Read_float32(&got, 3, true)
		if err != nil {
			t.Errorf("Test Read_float32 failed. err:%s\n", err)
		}

		if want != got {
			t.Errorf("Test Write_float32 failed. want:%v, got:%v\n", want, got)
		}
	}
}

func TestBuffer_float64(t *testing.T) {
	got := float64(0)
	for i := 0; i < 500; i++ {
		writer := NewBuffer()
		want := rand.Float64()

		err := writer.Write_float64(want, 3)
		if err != nil {
			t.Errorf("Test Write_float64 failed. err:%s\n", err)
		}

		reader := r(writer)
		err = reader.Read_float64(&got, 3, true)
		if err != nil {
			t.Errorf("Test Read_float64 failed. err:%s\n", err)
		}

		if want != got {
			t.Errorf("Test Write_float64 failed. want:%v, got:%v\n", want, got)
		}
	}
}

func TestBuffer_getTypeStr(t *testing.T) {
	wants := []string{
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
		"invalidType",
	}
	for k, want := range wants {
		got := getTypeStr(k)
		if got != want {
			t.Errorf("Test getTypeStr failed. want:%v, got:%v\n", want, got)
		}
	}
}


func TestReader_Reset(t *testing.T) {
	writer := NewBuffer()
	err := writer.Write_bytes([]byte("test"))
	if err != nil {
		t.Errorf("Write bytes failed")
	}
	reader := r(writer)
	reader.Reset([]byte{})
	writer.Reset()

	if writer.buf.String() != reader.ToString() {
		t.Errorf("Test Reset failed. want:%q, got:%q\n",
			writer.buf.String(), reader.ToString())
	}
}

func TestReader_Skip(t *testing.T) {
	writer := NewBuffer()
	err := writer.Write_bytes([]byte("hellotars"))
	if err != nil {
		t.Errorf("Write bytes failed")
	}
	reader := r(writer)
	reader.Skip(5)
	got := make([]byte, 4)
	reader.Read_bytes(&got, 4, true)
	if string(got) != "tars" {
		t.Errorf("Test Skip failed. want:%q, got:%q\n", "tars", string(got))
	}
}

func TestReader_ToBytes(t *testing.T) {
	writer := NewBuffer()
	want := []byte("hellotars")
	err := writer.Write_bytes(want)
	if err != nil {
		t.Errorf("Write bytes failed")
	}
	reader := r(writer)
	got := reader.ToBytes()
	if string(got) != string(want) {
		t.Errorf("Test reader ToBytes failed. want:%v, got:%v\n", want, got)
	}
}

func TestReader_unreadHead(t *testing.T) {
	writer := NewBuffer()
	err := writer.Write_string("hello", 0)
	if err != nil {
		t.Errorf("Write buffer failed.err:%v\n", err)
	}
	err = writer.Write_uint8(1, 1)
	if err != nil {
		t.Errorf("Write buffer failed.err:%v\n", err)
	}
	err = writer.Write_float32(1.2, 2)
	if err != nil {
		t.Errorf("Write buffer failed.err:%v\n", err)
	}

	// string type read head
	reader := r(writer)
	wantType, wantTag := STRING1, byte(0)
	gotType, gotTag, err := reader.readHead()
	if err != nil {
		t.Errorf("Read buffer failed.err:%v\n", err)
	}
	if gotType != wantType || gotTag != wantTag {
		t.Errorf("Failed to readHead. wantType:%v, wantTag:%v, gotType:%v, gotTag:%v\n",
			wantType, wantTag, gotType, gotTag)
	}

	// string type unread head
	reader.unreadHead(gotTag)
	gotType, gotTag, err = reader.readHead()
	// skip next 6 byte. 1 byte for string length, 5 byte for string itself.
	reader.Skip(6)
	if gotType != wantType || gotTag != wantTag {
		t.Errorf("Failed to readHead. wantType:%v, wantTag:%v, gotType:%v, gotTag:%v\n",
			wantType, wantTag, gotType, gotTag)
	}

	// uint8 read head
	wantType, wantTag = BYTE, byte(1)
	gotType, gotTag, err = reader.readHead()
	if err != nil {
		t.Errorf("Read buffer failed.err:%v\n", err)
	}
	if gotType != wantType || gotTag != wantTag {
		t.Errorf("Failed to readHead. wantType:%v, wantTag:%v, gotType:%v, gotTag:%v\n",
			wantType, wantTag, gotType, gotTag)
	}

	// uint8 unread head
	reader.unreadHead(gotTag)
	gotType, gotTag, err = reader.readHead()
	if gotType != wantType || gotTag != wantTag {
		t.Errorf("Failed to readHead. wantType:%v, wantTag:%v, gotType:%v, gotType:%v\n",
			wantType, wantTag, gotType, gotTag)
	}
}

func TestReader_SkipToNoCheck(t *testing.T) {
	prepareWrite := func() *Buffer {
		writer := NewBuffer()
		err := writer.Write_string("hello", 0)
		if err != nil {
			t.Errorf("Write buffer failed.err:%v\n", err)
		}
		err = writer.Write_uint8(1, 1)
		if err != nil {
			t.Errorf("Write buffer failed.err:%v\n", err)
		}
		err = writer.Write_float32(1.2, 2)
		if err != nil {
			t.Errorf("Write buffer failed.err:%v\n", err)
		}
		err = writer.Write_bool(true, 5)
		if err != nil {
			t.Errorf("Write buffer failed.err:%v\n", err)
		}
		return writer
	}

	reader := r(prepareWrite())
	err, exists, _ := reader.SkipToNoCheck(3, true)
	if err == nil || exists{
		t.Error("SkipToNoCheck failed.expecting error, but got nil\n")
	}
	if err != nil && err.Error() != "Can not find Tag 3. But require. tagCur: 5, tyCur: 0" {
		t.Errorf("SkipToNoCheck failed.expecting:%q, but got:%q\n",
			"Can not find Tag 3. But require. tagCur: 5, tyCur: 0", err)
	}

	reader = r(prepareWrite())
	err, exists, gotType := reader.SkipToNoCheck(2, true)
	if err != nil || !exists {
		t.Errorf("SkipToNoCheck failed.expecting nil error, but got:%v\n", err)
	}
	if gotType != FLOAT {
		t.Errorf("SkipToNoCheck error. wantType;%v, gotType:%v \n", FLOAT, gotType)
	}
}