package firmeware

import (
	"github.com/coorify/be/device"
	"github.com/coorify/be/modbus"
)

func Version(driver *device.Driver) uint16 {
	mdb := modbus.New(driver, modbus.NewRTUParser())

	mdb.Open()
	defer mdb.Close()

	req := modbus.NewRequest(modbus.OPCODE_READ_HOLDING_REGISTERS)
	req.Address = 0
	req.SetLength(1)

	rep := mdb.Exec(1, req)
	if rep == nil {
		return 0
	}

	pyd := rep.Payload()
	if pyd == nil {
		return 0
	}

	pydU16 := pyd.(modbus.PayloadU16)
	return pydU16.Get(0)
}
