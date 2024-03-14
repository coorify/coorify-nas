package target

type _esp32c3 struct {
}

func (e *_esp32c3) GetEraseSize(addr uint32, size uint32) uint32 {
	return size
}

func (e *_esp32c3) ReadMac(l Loader) ([]byte, error) {
	mac0, err := l.ReadReg(0x60008844)
	if err != nil {
		return nil, err
	}

	mac1, err := l.ReadReg(0x60008848)
	if err != nil {
		return nil, err
	}

	var macs [6]byte
	macs[0] = byte(mac1 >> 8)
	macs[1] = byte(mac1 >> 0)
	macs[2] = byte(mac0 >> 24)
	macs[3] = byte(mac0 >> 16)
	macs[4] = byte(mac0 >> 8)
	macs[5] = byte(mac0 >> 0)

	return macs[:], nil
}

func (e *_esp32c3) StubText(l Loader) (uint32, []byte) {
	bytes, err := l.ReadFile("embed/stub/esp32c3_text.bin")
	if err != nil {
		panic(err)
	}
	return 0x40380000, bytes
}

func (e *_esp32c3) StubData(l Loader) (uint32, []byte) {
	bytes, err := l.ReadFile("embed/stub/esp32c3_data.bin")
	if err != nil {
		panic(err)
	}
	return 0x3FC96BB0, bytes
}

func (e *_esp32c3) StubEntry() uint32 {
	return 0x4038069C
}
