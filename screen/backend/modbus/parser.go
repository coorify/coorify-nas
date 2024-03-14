package modbus

type Parser interface {
	Encode(addr *uint8, req *Request, b Buffer) bool
	Decode(addr *uint8, rep *Reply, b Buffer) bool
}

type _rtuParser struct {
}

func rtu_crc16(b Buffer, max int) uint16 {
	table := []uint16{0x0000, 0xA001}
	crc := uint16(0xFFFF)
	bit := uint8(0x00)
	xor := uint32(0x00000000)
	len := b.Length()
	val := uint8(0)

	if len > max {
		len = max
	}

	for len > 0 {
		b.ReadU8(&val)

		crc ^= uint16(val)
		for bit = 0; bit < 8; bit++ {
			xor = uint32(crc) & 1
			crc >>= 1
			crc ^= table[xor]
		}

		len -= 1
	}

	return crc
}

func NewRTUParser() Parser {
	return &_rtuParser{}
}

func (p *_rtuParser) Encode(addr *uint8, req *Request, b Buffer) bool {

	b.WriteU8(*addr)
	b.WriteU8(req.opcode)
	b.WriteU16(req.Address, true)
	b.WriteU16(req.lenOrVal, true)

	if opcodeRequestHaspayload(req.opcode) {
		if !req.payload.WriteTo(b) {
			return false
		}
	}

	crc_reader := b.Clone()
	crc16 := rtu_crc16(crc_reader, crc_reader.Length())

	return b.WriteU16(crc16, false)
}

func (p *_rtuParser) Decode(addr *uint8, rep *Reply, b Buffer) bool {
	for {
		reader := b.Clone()
		crc_reader := reader.Clone()

		if !reader.ReadU8(addr) {
			return false
		}

		if !reader.ReadU8(&rep.opcode) {
			return false
		}

		if opcodeReplyHasAttr(rep.opcode) {
			if !reader.ReadU16(&rep.Address, true) {
				return false
			}

			if !reader.ReadU16(&rep.lenOrVal, true) {
				return false
			}
		}

		if opcodeReplyHasPayload(rep.opcode) {
			if opcodeHasErr(rep.opcode) {
				rep.lenOrVal = 1
			} else {
				l := uint8(0)
				if !reader.ReadU8(&l) {
					return false
				}
				rep.lenOrVal = uint16(l)
			}

			if opcodeReplyPayloadBit(rep.opcode) {
				payload := &_payloadBit{}
				payload.SetLength(int(rep.lenOrVal) * 8)
				reader.Read(payload.raws)
				rep.payload = payload
			}

			if opcodeReplyPayloadU16(rep.opcode) {
				payload := &_payloadU16{}
				payload.SetLength(int(rep.lenOrVal) / 2)
				reader.Read(payload.raws)
				rep.payload = payload
			}
		}

		crc := uint16(0)
		max := crc_reader.Length() - reader.Length()

		if !reader.ReadU16(&crc, false) {
			return false
		}

		if rtu_crc16(crc_reader, max) == crc {
			return true
		} else {
			b.ReadU8(addr)
		}
	}
}
