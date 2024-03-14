package modbus

import (
	"io"
)

const (
	BUFFER_EMPTY = uint8(0x01)
	BUFFER_FULL  = uint8(0x02)
)

type Buffer interface {
	IsEmpty() bool
	IsFull() bool

	Free() int
	Length() int
	Clone() Buffer

	WriteU8(v uint8) bool
	WriteU16(v uint16, msb bool) bool
	Write(v []uint8) int
	WriteTo(w io.Writer) int

	ReadU8(v *uint8) bool
	ReadU16(v *uint16, msb bool) bool
	Read(v []uint8) int
	ReadFrom(r io.Reader) int
}

type _buffer struct {
	capacity int
	writpos  int
	readpos  int
	flag     uint8
	raws     []uint8
}

func (b *_buffer) IsEmpty() bool {
	return (b.flag & BUFFER_EMPTY) == BUFFER_EMPTY
}

func (b *_buffer) IsFull() bool {
	return (b.flag & BUFFER_FULL) == BUFFER_FULL
}

func (b *_buffer) Free() int {
	return b.capacity - b.Length()
}

func (b *_buffer) Length() int {
	if b.IsEmpty() {
		return 0
	}

	if b.IsFull() {
		return b.capacity
	}

	var len int
	len = 0
	len += b.writpos - b.readpos
	len += b.capacity
	len %= b.capacity
	return len
}

func (b *_buffer) WriteU8(v uint8) bool {
	return b.Write([]uint8{v}) == 1
}

func (b *_buffer) WriteU16(v uint16, msb bool) bool {
	var v8 []uint8
	if msb {
		v8 = []uint8{(uint8(v >> 8)), uint8(v)}
	} else {
		v8 = []uint8{(uint8(v)), uint8(v >> 8)}
	}

	return b.Write(v8) == 2
}

func (b *_buffer) Write(v []uint8) int {
	writed := 0
	vlen := len(v)

	if b.Free() < vlen {
		return writed
	}

	write_pos := b.raws[b.writpos:]
	write_len := 0
	if b.readpos > b.writpos {
		write_len = b.readpos - b.writpos
	} else {
		write_len = b.capacity - b.writpos
	}

	if write_len > vlen {
		write_len = vlen
	}

	copy(write_pos, v)
	b.writpos += write_len
	b.writpos %= b.capacity
	b.flag &= ^BUFFER_EMPTY
	if b.writpos == b.readpos {
		b.flag |= BUFFER_FULL
	}

	writed += write_len
	vlen -= write_len
	v = v[write_len:]
	if vlen > 0 {
		writed += b.Write(v)
	}

	return writed
}

func (b *_buffer) WriteTo(w io.Writer) int {
	writed := 0
	if b.IsEmpty() {
		return writed
	}

	read_pos := b.raws[b.readpos:]
	read_len := 0
	if b.writpos > b.readpos {
		read_len = b.writpos - b.readpos
	} else {
		read_len = b.capacity - b.readpos
	}

	write_len, _ := w.Write(read_pos[0:read_len])
	if write_len == 0 {
		return writed
	}

	b.readpos += write_len
	b.readpos %= b.capacity
	b.flag &= ^BUFFER_FULL
	if b.readpos == b.writpos {
		b.flag |= BUFFER_EMPTY
	}

	writed += write_len
	if read_len == write_len {
		writed += b.WriteTo(w)
	}

	return writed
}

func (b *_buffer) Read(v []uint8) int {
	readed := 0
	vlen := len(v)

	if b.IsEmpty() {
		return readed
	}

	read_pos := b.raws[b.readpos:]
	read_len := 0
	if b.writpos > b.readpos {
		read_len = b.writpos - b.readpos
	} else {
		read_len = b.capacity - b.readpos
	}

	if read_len > vlen {
		read_len = vlen
	}

	copy(v, read_pos)
	b.readpos += read_len
	b.readpos %= b.capacity
	b.flag &= ^BUFFER_FULL
	if b.readpos == b.writpos {
		b.flag |= BUFFER_EMPTY
	}

	readed += read_len
	vlen -= read_len
	v = v[read_len:]
	if vlen > 0 {
		readed += b.Read(v)
	}
	return readed
}

func (b *_buffer) ReadFrom(r io.Reader) int {
	writed := 0

	if b.IsFull() {
		return writed
	}

	write_pos := b.raws[b.writpos:]
	write_len := 0
	if b.readpos > b.writpos {
		write_len = b.readpos - b.writpos
	} else {
		write_len = b.capacity - b.writpos
	}

	read_len, _ := r.Read(write_pos[0:write_len])
	if read_len == 0 {
		return writed
	}

	b.writpos += read_len
	b.writpos %= b.capacity
	b.flag &= ^BUFFER_EMPTY
	if b.writpos == b.readpos {
		b.flag |= BUFFER_FULL
	}

	writed += read_len
	if read_len == write_len {
		writed += b.ReadFrom(r)
	}

	return writed
}

func (b *_buffer) ReadU8(v *uint8) bool {
	buf := []uint8{0}
	n := b.Read(buf)
	*v = buf[0]
	return n == 1
}

func (b *_buffer) ReadU16(v *uint16, msb bool) bool {
	v8 := []uint8{0, 0}
	if b.Read(v8) != 2 {
		return false
	}

	if msb {
		*v = (uint16(v8[0]) << 8) | uint16(v8[1])
	} else {
		*v = (uint16(v8[1]) << 8) | uint16(v8[0])
	}

	return true
}

func (b *_buffer) Clone() Buffer {
	return &_buffer{
		capacity: b.capacity,
		writpos:  b.writpos,
		readpos:  b.readpos,
		flag:     b.flag,
		raws:     b.raws,
	}
}

func NewBuffer(capacity int) Buffer {
	return &_buffer{
		capacity: capacity,
		readpos:  0,
		writpos:  0,
		flag:     BUFFER_EMPTY,
		raws:     make([]uint8, capacity),
	}
}

func NewBufferWriter(src []byte, capacity int) Buffer {
	return &_buffer{
		capacity: capacity,
		readpos:  0,
		writpos:  0,
		flag:     BUFFER_EMPTY,
		raws:     src,
	}
}

func NewBufferWith(src []byte, capacity int) Buffer {
	return &_buffer{
		capacity: capacity,
		readpos:  0,
		writpos:  0,
		flag:     BUFFER_EMPTY,
		raws:     src,
	}
}
