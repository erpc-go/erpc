package jce

import (
	"bufio"
	"encoding/binary"
	"io"
	"math"
	"unsafe"
)

// Encoder is wrapper of bytes.Encoder
type Encoder struct {
	buf *bufio.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		buf: bufio.NewWriter(w),
	}
}

//go:nosplit
func (w *Encoder) writeU8(data uint8) (err error) {
	return w.buf.WriteByte(data)
}

//go:nosplit
func (w *Encoder) writeU16(data uint16) (err error) {
	var (
		b  [2]byte
		bs []byte
	)

	bs = b[:]
	binary.BigEndian.PutUint16(bs, data)

	_, err = w.buf.Write(bs)
	return
}

//go:nosplit
func (w *Encoder) writeU32(data uint32) (err error) {
	var (
		b  [4]byte
		bs []byte
	)

	bs = b[:]
	binary.BigEndian.PutUint32(bs, data)

	_, err = w.buf.Write(bs)
	return
}

//go:nosplit
func (w *Encoder) writeU64(data uint64) (err error) {
	var (
		b  [8]byte
		bs []byte
	)

	bs = b[:]
	binary.BigEndian.PutUint64(bs, data)

	_, err = w.buf.Write(bs)
	return
}

//go:nosplit
func (b *Encoder) WriteHead(ty byte, tag byte) (err error) {
	if tag < 15 {
		return b.buf.WriteByte((tag << 4) | ty)
	}

	if err = b.buf.WriteByte((15 << 4) | ty); err != nil {
		return
	}

	return b.buf.WriteByte(tag)
}

// WriteSliceUint8 write []uint8 to the buffer.
func (b *Encoder) WriteSliceUint8(data []uint8) (err error) {
	_, err = b.buf.Write(data)
	return
}

// WriteSliceInt8 write []int8 to the buffer.
func (b *Encoder) WriteSliceInt8(data []int8) (err error) {
	_, err = b.buf.Write(*(*[]uint8)(unsafe.Pointer(&data)))
	return
}

// WriteBytes write []byte to the buffer
func (b *Encoder) WriteBytes(data []byte) (err error) {
	_, err = b.buf.Write(data)
	return
}

// WriteInt8 write int8 with the tag.
func (b *Encoder) WriteInt8(data int8, tag byte) (err error) {
	if data == 0 {
		return b.WriteHead(ZeroTag, tag)
	}

	if err = b.WriteHead(BYTE, tag); err != nil {
		return
	}

	if err = b.buf.WriteByte(byte(data)); err != nil {
		return
	}

	return
}

// WriteUint8 write uint8 with the tag
func (b *Encoder) WriteUint8(data uint8, tag byte) (err error) {
	return b.WriteInt16(int16(data), tag)
}

// WriteBool write bool with the tag.
func (b *Encoder) WriteBool(data bool, tag byte) (err error) {
	tmp := int8(0)
	if data {
		tmp = 1
	}
	return b.WriteInt8(tmp, tag)
}

// WriteInt16 write the int16 with the tag.
func (b *Encoder) WriteInt16(data int16, tag byte) (err error) {
	if data >= math.MinInt8 && data <= math.MaxInt8 {
		return b.WriteInt8(int8(data), tag)
	}
	if err = b.WriteHead(SHORT, tag); err != nil {
		return
	}

	return b.writeU16(uint16(data))
}

// WriteUint16 write uint16 with the tag.
func (b *Encoder) WriteUint16(data uint16, tag byte) (err error) {
	return b.WriteInt32(int32(data), tag)
}

// WriteInt32 write int32 with the tag.
func (b *Encoder) WriteInt32(data int32, tag byte) (err error) {
	if data >= math.MinInt16 && data <= math.MaxInt16 {
		return b.WriteInt16(int16(data), tag)
	}
	if err = b.WriteHead(INT, tag); err != nil {
		return
	}

	if err = b.writeU32(uint32(data)); err != nil {
		return
	}
	return
}

// WriteUint32 write uint32 data with the tag.
func (b *Encoder) WriteUint32(data uint32, tag byte) (err error) {
	return b.WriteInt64(int64(data), tag)
}

// WriteInt64 write int64 with the tag.
func (b *Encoder) WriteInt64(data int64, tag byte) (err error) {
	if data >= math.MinInt32 && data <= math.MaxInt32 {
		return b.WriteInt32(int32(data), tag)
	}

	if err = b.WriteHead(LONG, tag); err != nil {
		return
	}

	if err = b.writeU64(uint64(data)); err != nil {
		return
	}

	return
}

// WriteFloat32 writes float32 with the tag.
func (b *Encoder) WriteFloat32(data float32, tag byte) (err error) {
	if err = b.WriteHead(FLOAT, tag); err != nil {
		return err
	}

	return b.writeU32(math.Float32bits(data))
}

// WriteFloat64 writes float64 with the tag.
func (b *Encoder) WriteFloat64(data float64, tag byte) (err error) {
	if err = b.WriteHead(DOUBLE, tag); err != nil {
		return
	}

	return b.writeU64(math.Float64bits(data))
}

// WriteString writes string data with the tag.
func (b *Encoder) WriteString(data string, tag byte) (err error) {
	if len(data) > 255 {
		if err = b.WriteHead(STRING4, tag); err != nil {
			return err
		}

		if err = b.writeU32(uint32(len(data))); err != nil {
			return err
		}
	} else {
		if err = b.WriteHead(STRING1, tag); err != nil {
			return err
		}

		if err = b.writeU8(byte(len(data))); err != nil {
			return err
		}
	}

	_, err = b.buf.WriteString(data)

	return
}
