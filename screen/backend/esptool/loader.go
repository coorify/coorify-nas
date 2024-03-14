package esptool

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"io/fs"
	"net"
	"time"

	"github.com/coorify/be/esptool/target"
	"github.com/sirupsen/logrus"
)

type Driver interface {
	Open() error
	Close() error

	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

const (
	ESPOP_FLASHBEGIN     = 0x02
	ESPOP_MEMBEGIN       = 0x05
	ESPOP_MEMEND         = 0x06
	ESPOP_MEMDATA        = 0x07
	ESPOP_SYNC           = 0x08
	ESPOP_READREG        = 0x0a
	ESPOP_FLASHDEFLBEGIN = 0x10
	ESPOP_FLASHDEFLDATA  = 0x11
	ESPOP_FLASHDEFLEND   = 0x12
	ESPOP_ERASEFLASH     = 0xd0

	ESP_RAMBLOCK   = 0x1800
	ESP_FLASHBLOCK = 0x400
)

type Loader struct {
	efs fs.FS
	drv Driver
	rom target.ROM
}

func NewLoader(drv Driver, embedFS fs.FS) *Loader {
	return &Loader{
		drv: drv,
		efs: embedFS,
	}
}

func (l *Loader) Close() error {
	return l.drv.Close()
}

func (l *Loader) Open() error {
	if err := l.drv.Open(); err != nil {
		return err
	}

	if err := l.Sync(10); err != nil {
		return err
	}

	val, err := l.ReadReg(0x40001000)
	if err != nil {
		return err
	}

	l.rom = target.MagicToRom(val)
	if l.rom == nil {
		return fmt.Errorf("esptool: chip not support")
	}

	mac, err := l.ReadMac()
	if err != nil {
		return err
	}
	logrus.Infof("esptool: chip mac(%s)", mac)

	return l.RunStub()
}

func (l *Loader) Sync(retryMax int) error {
	pkt := []byte{0x07, 0x07, 0x12, 0x20}
	pkt = append(pkt, bytes.Repeat([]byte{0x55}, 32)...)

	var err error

	for retry := 0; retry < retryMax; retry++ {
		_, _, err = l.exec(ESPOP_SYNC, pkt, 0, time.Second)
		if err == nil {
			return nil
		}
	}

	return fmt.Errorf("esptool: sync error")
}

func (l *Loader) ReadMac() (string, error) {
	mac, err := l.rom.ReadMac(l)
	if err != nil {
		return "", err
	}

	return net.HardwareAddr(mac).String(), nil
}

func (l *Loader) ReadReg(addr uint32) (uint32, error) {
	pkt := uint32ToBytes(addr)
	val, _, err := l.exec(ESPOP_READREG, pkt, 0, time.Second)
	return val, err
}

func (l *Loader) ReadFile(name string) ([]byte, error) {
	f, err := l.efs.Open(name)
	if err != nil {
		return make([]byte, 0), err
	}

	defer f.Close()
	return io.ReadAll(f)
}

func (l *Loader) EraseFlash() error {
	logrus.Info("esptool: erasing flash (this may take a while)...")
	_, _, err := l.exec(ESPOP_ERASEFLASH, make([]byte, 0), 0, 20*time.Second)
	if err == nil {
		logrus.Info("esptool: chip erase completed successfully")
	}
	return err
}

func (l *Loader) WriteFlash(addr uint32, image []byte) error {
	var b bytes.Buffer
	w, _ := zlib.NewWriterLevel(&b, 9)
	_, _ = w.Write(image)
	w.Close()

	zimage := b.Bytes()

	limg := len(image)
	lzimg := len(zimage)
	// numBlocks := (uint32(len(image)) + ESP_FLASHBLOCK - 1) / ESP_FLASHBLOCK
	znumBlocks := (lzimg + ESP_FLASHBLOCK - 1) / ESP_FLASHBLOCK
	logrus.Infof("esptool: compressed %d bytes to %d bytes. Ration = %.1f", limg, lzimg, float64(lzimg)/float64(limg))

	if err := l.FlashDeflBegin(uint32(limg), uint32(znumBlocks), ESP_FLASHBLOCK, addr); err != nil {
		return err
	}

	if err := l.block(ESPOP_FLASHDEFLDATA, uint32(znumBlocks), ESP_FLASHBLOCK, zimage); err != nil {
		return err
	}

	return nil
}

func (l *Loader) WriteFlashFinish() error {
	if err := l.FlashBegin(0, 0); err != nil {
		return err
	}

	if err := l.FlashDeflFinish(false); err != nil {
		return err
	}

	return nil
}

func (l *Loader) FlashDeflBegin(eraseSize uint32, numBlocks uint32, blockSize uint32, offset uint32) error {
	pkt := make([]byte, 0)
	pkt = append(pkt, uint32ToBytes(eraseSize)...)
	pkt = append(pkt, uint32ToBytes(numBlocks)...)
	pkt = append(pkt, uint32ToBytes(blockSize)...)
	pkt = append(pkt, uint32ToBytes(offset)...)

	_, _, err := l.exec(ESPOP_FLASHDEFLBEGIN, pkt, 0, time.Second)
	return err
}

func (l *Loader) FlashDeflFinish(reboot bool) error {
	isReboot := 1
	if reboot {
		isReboot = 0
	}

	pkt := make([]byte, 0)
	pkt = append(pkt, uint32ToBytes(uint32(isReboot))...)
	_, _, err := l.exec(ESPOP_FLASHDEFLEND, pkt, 0, time.Second)
	return err
}

func (l *Loader) FlashBegin(size uint32, addr uint32) error {
	numBlocks := (size + ESP_FLASHBLOCK - 1) / ESP_FLASHBLOCK
	eraseSize := l.rom.GetEraseSize(addr, size)

	pkt := make([]byte, 0)
	pkt = append(pkt, uint32ToBytes(eraseSize)...)
	pkt = append(pkt, uint32ToBytes(numBlocks)...)
	pkt = append(pkt, uint32ToBytes(ESP_FLASHBLOCK)...)
	pkt = append(pkt, uint32ToBytes(addr)...)

	_, _, err := l.exec(ESPOP_FLASHBEGIN, pkt, 0, time.Second)
	return err
}

func (l *Loader) MemBegin(size uint32, blocks uint32, blocksize uint32, offset uint32) error {
	pkt := make([]byte, 0)
	pkt = append(pkt, uint32ToBytes(size)...)
	pkt = append(pkt, uint32ToBytes(blocks)...)
	pkt = append(pkt, uint32ToBytes(blocksize)...)
	pkt = append(pkt, uint32ToBytes(offset)...)

	_, _, err := l.exec(ESPOP_MEMBEGIN, pkt, 0, time.Second)
	return err
}

func (l *Loader) MemBlock(blocks uint32, blocksize uint32, bytes []byte) error {
	return l.block(ESPOP_MEMDATA, blocks, blocksize, bytes)
}

func (l *Loader) MemFinish(entry uint32) error {
	isEntry := 0
	if entry == 0 {
		isEntry = 1
	}

	pkt := make([]byte, 0)
	pkt = append(pkt, uint32ToBytes(uint32(isEntry))...)
	pkt = append(pkt, uint32ToBytes(entry)...)
	_, _, err := l.exec(ESPOP_MEMEND, pkt, 0, time.Second)
	return err
}

func (l *Loader) RunStub() error {
	var addr, blen, blocks uint32
	var raws []byte

	logrus.Info("esptool: uploading stub...")

	addr, raws = l.rom.StubText(l)
	blen = uint32(len(raws))
	blocks = (blen + ESP_RAMBLOCK - 1) / ESP_RAMBLOCK

	if err := l.MemBegin(blen, uint32(blocks), ESP_RAMBLOCK, addr); err != nil {
		return err
	}
	if err := l.MemBlock(blocks, ESP_RAMBLOCK, raws); err != nil {
		return err
	}

	addr, raws = l.rom.StubData(l)
	blen = uint32(len(raws))
	blocks = (blen + ESP_RAMBLOCK - 1) / ESP_RAMBLOCK
	if err := l.MemBegin(blen, uint32(blocks), ESP_RAMBLOCK, addr); err != nil {
		return err
	}
	if err := l.MemBlock(blocks, ESP_RAMBLOCK, raws); err != nil {
		return err
	}

	logrus.Info("esptool: running stub...")
	addr = l.rom.StubEntry()

	if err := l.MemFinish(addr); err != nil {
		return nil
	}

	time.Sleep(100 * time.Microsecond)

	res, err := SlipRead(l.drv, 10*time.Second)
	if err != nil {
		return err
	}

	if res[0] == 79 && res[1] == 72 && res[2] == 65 && res[3] == 73 {
		logrus.Info("esptool: stub running...")
		return nil
	}

	return fmt.Errorf("esptool: failed to start stub")
}

func (l *Loader) exec(op byte, data []byte, cks uint32, timeout time.Duration) (uint32, []byte, error) {
	if data == nil {
		data = make([]byte, 0)
	}

	dlen := len(data)
	lBytes := uint16ToBytes(uint16(dlen))
	cBytes := uint32ToBytes(cks)

	pkt := make([]byte, 8+dlen)
	pkt[0] = 0x00
	pkt[1] = op
	copy(pkt[2:], lBytes)
	copy(pkt[4:], cBytes)
	copy(pkt[8:], data)

	logrus.Tracef("command op:0x%02X data len=%d data=%s", op, dlen, hexConvert(data))

	err := SlipWrite(l.drv, pkt)
	if err != nil {
		return 0, nil, err
	}

	for retryCount := 0; retryCount < 16; retryCount++ {
		replyBytes, err := SlipRead(l.drv, timeout)

		if err != nil {
			return 0, nil, err
		}

		if replyBytes[1] != byte(op) {
			continue
		} else {
			return bytesToUint32(replyBytes[4:8]), replyBytes[8:], nil
		}
	}

	return 0, nil, fmt.Errorf("esptool: slip timeout")
}

func (l *Loader) block(op byte, blocks uint32, blocksize uint32, bytes []byte) error {
	sequence := uint32(0)
	sent := uint32(0)
	total := uint32(len(bytes))

	for {
		logrus.Debugf("esptool: %d of %d - %.2f", sent, total, float64(sent)/float64(total)*100.0)

		if sent >= total {
			break
		}

		blockLen := uint32(total - sent)
		if blockLen > blocksize {
			blockLen = blocksize
		}
		block := bytes[sent : sent+blockLen]

		pkt := make([]byte, 0)
		pkt = append(pkt, uint32ToBytes(uint32(len(block)))...)
		pkt = append(pkt, uint32ToBytes(sequence)...)
		pkt = append(pkt, uint32ToBytes(0)...)
		pkt = append(pkt, uint32ToBytes(0)...)
		pkt = append(pkt, block...)
		_, _, err := l.exec(op, pkt, checksum(block), time.Second)

		if err != nil {
			return err
		}

		sequence++
		sent += blockLen
	}

	return nil
}
