package firmeware

import (
	"io"
	"io/fs"
	"time"

	"github.com/coorify/be/device"
	"github.com/coorify/be/esptool"
	"github.com/coorify/be/option"
	"github.com/sirupsen/logrus"
)

func download(driver *device.Driver, efs fs.FS) error {
	loader := esptool.NewLoader(driver, efs)
	if err := loader.Open(); err != nil {
		return err
	}
	defer loader.Close()

	if err := loader.EraseFlash(); err != nil {
		return err
	}

	var raws []byte
	f, _ := efs.Open("embed/bootloader.bin")
	raws, _ = io.ReadAll(f)
	if err := loader.WriteFlash(0x00000000, raws); err != nil {
		return err
	}
	f.Close()

	f, _ = efs.Open("embed/nas-ui.bin")
	raws, _ = io.ReadAll(f)
	if err := loader.WriteFlash(0x00010000, raws); err != nil {
		return err
	}
	f.Close()

	f, _ = efs.Open("embed/partition-table.bin")
	raws, _ = io.ReadAll(f)
	if err := loader.WriteFlash(0x00008000, raws); err != nil {
		return err
	}
	f.Close()

	if err := loader.WriteFlashFinish(); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)
	return nil
}

func Update(driver *device.Driver, o *option.UpdateOption) error {
	device.Reboot(driver, false)

	ever := o.Version
	hver := Version(driver)

	logrus.Infof("firmeware: hardware(%v) embedded(%v)", hver, ever)
	if hver == ever {
		return nil
	}

	logrus.Warnf("firmeware: update to %v", ever)
	device.Reboot(driver, true)

	if err := download(driver, o.EmbedFS); err != nil {
		return err
	}

	logrus.Warn("firmeware: update finished,reboot....")
	device.Reboot(driver, false)
	return nil
}
