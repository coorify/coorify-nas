package modbus

type Payload interface {
	IsBit() bool
	IsU16() bool

	SetLength(v int)
	WriteTo(b Buffer) bool
}

type PayloadBit interface {
	Payload

	Get(offset int) bool
	Set(offset int, val bool)
}

type PayloadU16 interface {
	Payload

	Get(offset int) uint16
	Set(offset int, val uint16)
}

type _payloadBit struct {
	raws []uint8
}

type _payloadU16 struct {
	raws []uint8
}

func (p *_payloadBit) IsBit() bool {
	return true
}

func (p *_payloadBit) IsU16() bool {
	return false
}

func (p *_payloadBit) SetLength(v int) {
	vlen := v / 8
	if v%8 != 0 {
		vlen += 1
	}

	p.raws = make([]uint8, vlen)
}

func (p *_payloadBit) WriteTo(b Buffer) bool {
	b.WriteU8(uint8(len(p.raws)))
	return b.Write(p.raws) == len(p.raws)
}

func (p *_payloadBit) Get(offset int) bool {
	pos := offset / 8
	bit := offset % 8

	return p.raws[pos]&(1<<uint8(bit)) != 0
}

func (p *_payloadBit) Set(offset int, val bool) {
	pos := offset / 8
	bit := offset % 8

	if val {
		p.raws[pos] |= (1 << uint8(bit))
	} else {
		p.raws[pos] &= ^(1 << uint8(bit))
	}
}

func (p *_payloadU16) IsBit() bool {
	return false
}

func (p *_payloadU16) IsU16() bool {
	return true
}

func (p *_payloadU16) SetLength(v int) {
	p.raws = make([]uint8, v*2)
}

func (p *_payloadU16) WriteTo(b Buffer) bool {
	b.WriteU8(uint8(len(p.raws)))
	return b.Write(p.raws) == len(p.raws)
}

func (p *_payloadU16) Get(offset int) uint16 {
	val := uint16(0)
	offset *= 2

	val |= uint16(p.raws[offset]) << 8
	val |= uint16(p.raws[offset+1])

	return val
}

func (p *_payloadU16) Set(offset int, val uint16) {
	offset *= 2
	p.raws[offset] = uint8(val >> 8)
	p.raws[offset+1] = uint8(val >> 0)
}
