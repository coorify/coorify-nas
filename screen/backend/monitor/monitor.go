package monitor

import (
	"context"
	"time"

	"github.com/coorify/be/device"
	"github.com/coorify/be/modbus"
	"github.com/coorify/be/openwrt"
)

type Monitor struct {
	wrt  openwrt.Client
	mdb  *modbus.Modbus
	exit context.CancelFunc
}

func NewMonitor(drv *device.Driver, wrt openwrt.Client) *Monitor {
	return &Monitor{
		wrt: wrt,
		mdb: modbus.New(drv, modbus.NewRTUParser()),
	}
}

func (m *Monitor) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
			m.metrics()
		}
	}
}

func (m *Monitor) metrics() {
	req := modbus.NewRequest(modbus.OPCODE_WRITE_REGISTERS)
	req.Address = 1
	req.SetLength(6)

	pyd := req.Payload().(modbus.PayloadU16)

	sys, err := m.wrt.SystemStatus()
	if err == nil {
		pyd.Set(0, sys.Cpu)
		pyd.Set(1, sys.Mem)
		pyd.Set(2, sys.Tmp)
	}

	sta, err := m.wrt.NetworkStatus()
	if err == nil {
		pyd.Set(3, sta.Up)
		pyd.Set(4, sta.Down)
		pyd.Set(5, sta.Num)
	}

	m.mdb.Exec(1, req)
}

func (m *Monitor) Start() error {
	if err := m.mdb.Open(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	m.exit = cancel

	go m.run(ctx)

	return nil
}

func (m *Monitor) Stop() error {
	m.exit()

	return m.mdb.Close()
}
