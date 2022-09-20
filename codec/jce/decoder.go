package jce

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"unsafe"
)

// Decoder is wrapper of bytes.Decoder
type Decoder struct {
	buf *bufio.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		buf: bufio.NewReader(r),
	}
}

//go:nosplit
func (r *Decoder) readU8(data *uint8) (err error) {
	*data, err = r.buf.ReadByte()
	return
}

//go:nosplit
func (r *Decoder) readU16(data *uint16) (err error) {
	var (
		b  [2]byte
		bs []byte
	)
	bs = b[:]
	_, err = r.buf.Read(bs)
	*data = binary.BigEndian.Uint16(bs)
	return
}

//go:nosplit
func (r *Decoder) readU32(data *uint32) (err error) {
	var (
		b  [4]byte
		bs []byte
	)
	bs = b[:]
	_, err = r.buf.Read(bs)
	*data = binary.BigEndian.Uint32(bs)
	return
}

//go:nosplit
func (r *Decoder) readU64(data *uint64) (err error) {
	var (
		b  [8]byte
		bs []byte
	)
	bs = b[:]
	_, err = r.buf.Read(bs)
	*data = binary.BigEndian.Uint64(bs)
	return
}

//go:nosplit
func (b *Decoder) readHead() (ty, tag byte, err error) {
	data, err := b.buf.ReadByte()
	if err != nil {
		return
	}

	ty = data & 0x0f
	tag = (data & 0xf0) >> 4
	if tag != 15 {
		return
	}

	data, err = b.buf.ReadByte()
	if err != nil {
		return
	}
	tag = data

	return
}

// unreadHead 回退一个head byte， curTag为当前读到的tag信息，当tag超过4位时则回退两个head byte
// unreadHead put back the current head byte.
func (b *Decoder) unreadHead(curTag byte) {
	_ = b.buf.UnreadByte()
	if curTag >= 15 {
		_ = b.buf.UnreadByte()
	}
}

// Next return the []byte of next n .
//
//go:nosplit
func (b *Decoder) Next(n int) []byte {
	res, err := b.buf.Peek(n)
	if err == nil {
		return []byte{}
	}
	return res
}

// Skip the next n byte.
//
//go:nosplit
func (b *Decoder) Skip(n int) {
	if _, err := b.buf.Discard(n); err != nil {
		fmt.Println(err)
	}
}

func (b *Decoder) skipFieldMap() error {
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

func (b *Decoder) skipFieldList() error {
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

func (b *Decoder) skipFieldSimpleList() error {
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

func (b *Decoder) skipField(ty byte) error {
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
		err := b.readU32(&l)
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
func (b *Decoder) SkipToStructEnd() error {
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
func (b *Decoder) SkipToNoCheck(tag byte, require bool) (bool, byte, error) {
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
func (b *Decoder) SkipTo(ty, tag byte, require bool) (bool, error) {
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
func (b *Decoder) ReadSliceInt8(data *[]int8, len int32, require bool) error {
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
func (b *Decoder) ReadSliceUint8(data *[]uint8, len int32, require bool) error {
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
func (b *Decoder) ReadBytes(data *[]byte, len int32, require bool) error {
	*data = make([]byte, len)
	_, err := b.buf.Read(*data)
	return err
}

// ReadInt8 reads the int8 data for the tag and the require or optional sign.
func (b *Decoder) ReadInt8(data *int8, tag byte, require bool) error {
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
		err = b.readU8(&tmp)
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
func (b *Decoder) ReadUint8(data *uint8, tag byte, require bool) error {
	n := int16(*data)
	err := b.ReadInt16(&n, tag, require)
	*data = uint8(n)
	return err
}

// ReadBool reads the bool value for the tag and the require or optional sign.
func (b *Decoder) ReadBool(data *bool, tag byte, require bool) error {
	var tmp int8
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
func (b *Decoder) ReadInt16(data *int16, tag byte, require bool) error {
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
		err = b.readU8(&tmp)
		*data = int16(int8(tmp))
	case SHORT:
		var tmp uint16
		err = b.readU16(&tmp)
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
func (b *Decoder) ReadUint16(data *uint16, tag byte, require bool) error {
	n := int32(*data)
	err := b.ReadInt32(&n, tag, require)
	*data = uint16(n)
	return err
}

// ReadInt32 reads the int32 value for the tag and the require or optional sign.
func (b *Decoder) ReadInt32(data *int32, tag byte, require bool) error {
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
		err = b.readU8(&tmp)
		*data = int32(int8(tmp))
	case SHORT:
		var tmp uint16
		err = b.readU16(&tmp)
		*data = int32(int16(tmp))
	case INT:
		var tmp uint32
		err = b.readU32(&tmp)
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
func (b *Decoder) ReadUint32(data *uint32, tag byte, require bool) error {
	n := int64(*data)
	err := b.ReadInt64(&n, tag, require)
	*data = uint32(n)
	return err
}

// ReadInt64 reads the int64 value for the tag and the require or optional sign.
func (b *Decoder) ReadInt64(data *int64, tag byte, require bool) error {
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
		err = b.readU8(&tmp)
		*data = int64(int8(tmp))
	case SHORT:
		var tmp uint16
		err = b.readU16(&tmp)
		*data = int64(int16(tmp))
	case INT:
		var tmp uint32
		err = b.readU32(&tmp)
		*data = int64(int32(tmp))
	case LONG:
		var tmp uint64
		err = b.readU64(&tmp)
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
func (b *Decoder) ReadFloat32(data *float32, tag byte, require bool) error {
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
		err = b.readU32(&tmp)
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
func (b *Decoder) ReadFloat64(data *float64, tag byte, require bool) error {
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
		err = b.readU32(&tmp)
		*data = float64(math.Float32frombits(tmp))
	case DOUBLE:
		var tmp uint64
		err = b.readU64(&tmp)
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
func (b *Decoder) ReadString(data *string, tag byte, require bool) (err error) {
	have, ty, err := b.SkipToNoCheck(tag, require)
	if err != nil {
		return
	}

	if !have {
		return
	}

	if ty == STRING4 {
		var length uint32
		if err = b.readU32(&length); err != nil {
			return fmt.Errorf("read string4 tag:%d error:%v", tag, err)
		}
		buff := b.Next(int(length))
		*data = string(buff)
		return
	}

	if ty == STRING1 {
		var length uint8
		if err = b.readU8(&length); err != nil {
			return fmt.Errorf("read string1 tag:%d error:%v", tag, err)
		}
		buff := b.Next(int(length))
		*data = string(buff)
		return
	}

	return fmt.Errorf("need string, tag:%d, but type is %s", tag, getTypeStr(int(ty)))
}
