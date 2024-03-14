package device

import (
	"errors"
	"time"

	"go.bug.st/serial"
)

var ErrPortNotOpen = errors.New("port not open")

type Driver struct {
	name string
	port serial.Port
	mode *serial.Mode
}

func NewDriver(name string) *Driver {
	return &Driver{
		name: name,
		port: nil,
		mode: &serial.Mode{
			BaudRate: 115200,
			DataBits: 8,
			Parity:   serial.NoParity,
			StopBits: serial.OneStopBit,
		},
	}
}

func (m *Driver) Open() error {
	var err error

	m.port, err = serial.Open(m.name, m.mode)
	if err != nil {
		return err
	}

	m.port.ResetInputBuffer()
	m.port.ResetOutputBuffer()
	m.port.SetReadTimeout(1 * time.Second)
	return nil
}

func (m *Driver) Close() error {
	if m.port == nil {
		return ErrPortNotOpen
	}

	return m.port.Close()
}

func (m *Driver) Read(p []byte) (int, error) {
	if m.port == nil {
		return 0, ErrPortNotOpen
	}

	return m.port.Read(p)
}

func (m *Driver) Write(p []byte) (int, error) {
	if m.port == nil {
		return 0, ErrPortNotOpen
	}

	return m.port.Write(p)
}

func (m *Driver) SetDTR(dtr bool) error {
	if m.port == nil {
		return ErrPortNotOpen
	}

	return m.port.SetDTR(dtr)
}

func (m *Driver) SetRTS(rts bool) error {
	if m.port == nil {
		return ErrPortNotOpen
	}

	return m.port.SetRTS(rts)
}
