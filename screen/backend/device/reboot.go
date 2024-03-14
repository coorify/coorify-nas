package device

import "time"

func Reboot(drv *Driver, download bool) {
	drv.Open()

	if download {
		drv.SetRTS(false)
		drv.SetDTR(false)
		time.Sleep(100 * time.Microsecond)
		drv.SetDTR(true)
		drv.SetRTS(false)
		time.Sleep(100 * time.Microsecond)
		drv.SetRTS(true)
		drv.SetRTS(false)
		drv.SetRTS(true)
		time.Sleep(100 * time.Microsecond)
		drv.SetDTR(false)
		drv.SetRTS(false)
	} else {
		drv.SetDTR(false)
		drv.SetRTS(true)
		time.Sleep(100 * time.Millisecond)
		drv.SetDTR(true)
		drv.SetRTS(false)
	}

	drv.Close()
	time.Sleep(5 * time.Second)
}
