package modbus

type Driver interface {
	Open() error
	Close() error

	Read([]byte) (int, error)
	Write([]byte) (int, error)
}
