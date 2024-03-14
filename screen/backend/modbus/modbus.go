package modbus

import (
	"fmt"
	"time"
)

type Modbus struct {
	driver Driver
	parser Parser
}

func New(driver Driver, parser Parser) *Modbus {

	return &Modbus{
		driver: driver,
		parser: parser,
	}
}

func (m *Modbus) Open() error {
	if err := m.driver.Open(); err != nil {
		return err
	}

	return nil
}

func (m *Modbus) Close() error {
	if err := m.driver.Close(); err != nil {
		return err
	}

	return nil
}

func (m *Modbus) Exec(addr uint8, req *Request) *Reply {
	byteLen := 256
	bytes := make([]byte, byteLen)

	writer := NewBufferWriter(bytes, byteLen)
	if !m.parser.Encode(&addr, req, writer) {
		return nil
	}
	writer.WriteTo(m.driver)

	time.Sleep(20 * time.Microsecond)

	reader := NewBufferWith(bytes, byteLen)
	reader.ReadFrom(m.driver)

	rep := &Reply{}
	if !m.parser.Decode(&addr, rep, reader) {
		return nil
	}
	return rep
}

func (m *Modbus) Test() {
	req := NewRequest(OPCODE_READ_HOLDING_REGISTERS)
	req.Address = 0
	req.SetLength(2)

	rep := m.Exec(1, req)
	pyd := rep.Payload().(PayloadU16)
	fmt.Printf("%v\n", pyd.IsU16())
	fmt.Printf("%v\n", pyd.Get(0))
}
