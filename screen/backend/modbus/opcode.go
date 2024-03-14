package modbus

const (
	OPCODE_READ_COILS             = 0x01
	OPCODE_DISCRETE_INPUTS        = 0x02
	OPCODE_READ_HOLDING_REGISTERS = 0x03
	OPCODE_READ_INPUT_REGISTERS   = 0x04
	OPCODE_WRITE_COIL             = 0x05
	OPCODE_WRITE_REGISTER         = 0x06
	OPCODE_WRITE_COILS            = 0x0F
	OPCODE_WRITE_REGISTERS        = 0x10
	OPCODE_ERROR_MASK             = 0x80
	OPCODE_FUNC_MASK              = 0x7F
)

func opcodeHasErr(opcode uint8) bool {
	return (opcode & OPCODE_ERROR_MASK) == OPCODE_ERROR_MASK
}

func opcodeMask(opcode uint8) uint8 {
	return (opcode & OPCODE_FUNC_MASK)
}

func opcodeAllowed(opcode uint8) bool {
	mopcode := opcodeMask(opcode)
	return (mopcode == OPCODE_READ_COILS) ||
		(mopcode == OPCODE_DISCRETE_INPUTS) ||
		(mopcode == OPCODE_READ_HOLDING_REGISTERS) ||
		(mopcode == OPCODE_READ_INPUT_REGISTERS) ||
		(mopcode == OPCODE_WRITE_COIL) ||
		(mopcode == OPCODE_WRITE_REGISTER) ||
		(mopcode == OPCODE_WRITE_COILS) ||
		(mopcode == OPCODE_WRITE_REGISTERS) ||
		(opcodeHasErr(opcode))
}

func opcodeRequestPayloadBit(opcode uint8) bool {
	return (opcode == OPCODE_WRITE_COILS)
}

func opcodeRequestPayloadU16(opcode uint8) bool {
	return (opcode == OPCODE_WRITE_REGISTERS)
}

func opcodeRequestHaspayload(opcode uint8) bool {
	return opcodeRequestPayloadBit(opcode) || opcodeRequestPayloadU16(opcode)
}

func opcodeReplyPayloadBit(opcode uint8) bool {
	return (opcode == OPCODE_READ_COILS) ||
		(opcode == OPCODE_DISCRETE_INPUTS) ||
		(opcodeHasErr(opcode))
}

func opcodeReplyPayloadU16(opcode uint8) bool {
	return (opcode == OPCODE_READ_HOLDING_REGISTERS) ||
		(opcode == OPCODE_READ_INPUT_REGISTERS)
}

func opcodeReplyHasPayload(opcode uint8) bool {
	return opcodeReplyPayloadBit(opcode) || opcodeReplyPayloadU16(opcode)
}

func opcodeReplyHasAttr(opcode uint8) bool {
	return (opcode == OPCODE_WRITE_COIL) ||
		(opcode == OPCODE_WRITE_REGISTER) ||
		(opcode == OPCODE_WRITE_COILS) ||
		(opcode == OPCODE_WRITE_REGISTERS)
}
