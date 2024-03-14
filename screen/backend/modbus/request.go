package modbus

type Request struct {
	Address  uint16
	opcode   uint8
	lenOrVal uint16
	payload  Payload
}

func NewRequest(opcode uint8) *Request {
	req := &Request{
		opcode:   opcode,
		lenOrVal: 0,
		payload:  nil,
	}

	if opcodeRequestPayloadBit(req.opcode) {
		req.payload = &_payloadBit{}
	}

	if opcodeRequestPayloadU16(req.opcode) {
		req.payload = &_payloadU16{}
	}

	return req
}

func (r *Request) OpCode() uint8 {
	return r.opcode
}

func (r *Request) SetLength(val uint16) {
	r.lenOrVal = val

	if r.payload != nil {
		r.payload.SetLength(int(val))
	}
}

func (r *Request) SetValue(val bool) {
	r.lenOrVal = 0x0000

	if val {
		r.lenOrVal = 0xFF00
	}
}

func (r *Request) Payload() Payload {
	return r.payload
}
