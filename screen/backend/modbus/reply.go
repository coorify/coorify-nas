package modbus

type Reply struct {
	Address  uint16
	opcode   uint8
	lenOrVal uint16
	payload  Payload
}

func (r *Reply) OpCode() uint8 {
	return r.opcode
}

func (r *Reply) Length() uint16 {
	return r.lenOrVal
}

func (r *Reply) Value() bool {
	return r.lenOrVal == 0xFF00
}

func (r *Reply) Payload() Payload {
	return r.payload
}
