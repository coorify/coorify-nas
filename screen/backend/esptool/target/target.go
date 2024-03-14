package target

type Loader interface {
	ReadReg(addr uint32) (uint32, error)
	ReadFile(name string) ([]byte, error)
}

type ROM interface {
	GetEraseSize(addr uint32, size uint32) uint32

	ReadMac(l Loader) ([]byte, error)

	StubText(l Loader) (uint32, []byte)
	StubData(l Loader) (uint32, []byte)
	StubEntry() uint32
}

func MagicToRom(magic uint32) ROM {

	switch magic {
	case 0x6921506f, 0x1b31506f, 0x4881606f, 0x4361606f:
		return &_esp32c3{}
	}

	return nil
}
