package esptool

import (
	"fmt"
	"io"
	"time"

	"github.com/sirupsen/logrus"
)

func SlipWrite(w io.Writer, data []byte) error {
	pkt := make([]byte, 0)

	pkt = append(pkt, 0xC0)
	for _, v := range data {
		if v == 0xDB {
			pkt = append(pkt, 0xDB, 0xDD)
		} else if v == 0xC0 {
			pkt = append(pkt, 0xDB, 0xDC)
		} else {
			pkt = append(pkt, v)
		}
	}
	pkt = append(pkt, 0xC0)

	logrus.Debugf("Write %d bytes: %s", len(pkt), hexConvert(pkt))

	plen := len(pkt)
	sent := 0
	for sent != plen {
		n, err := w.Write(pkt[sent:])
		if err != nil {
			return err
		}
		sent += n
	}

	return nil
}

func SlipRead(r io.Reader, timeout time.Duration) ([]byte, error) {
	const waitingForHeader byte = 0
	const readingContent byte = 1
	const inEscape byte = 2

	startTime := time.Now()
	bytes := make([]byte, 1)
	result := make([]byte, 0)
	state := waitingForHeader

	for {
		if time.Since(startTime) > timeout {
			return nil, fmt.Errorf("esptool: read timeout after %v. Received %d bytes", time.Since(startTime), len(result))
		}

		n, err := r.Read(bytes)
		if err != nil {
			if err.Error() == "EOF" {
				continue
			}
			return nil, err
		}

		if n != 1 {
			continue
		}

		switch state {
		case waitingForHeader:
			if bytes[0] == 0xC0 {
				state = readingContent
			}
		case readingContent:
			switch bytes[0] {
			case 0xC0:
				logrus.Debugf("Read %d bytes: %s", len(result), hexConvert(result))
				return result, nil
			case 0xDB:
				state = inEscape
			default:
				result = append(result, bytes[0])
			}
		case inEscape:
			switch bytes[0] {
			case 0xDC:
				result = append(result, 0xC0)
				state = readingContent
			case 0xDD:
				result = append(result, 0xDB)
				state = readingContent
			default:
				return nil, fmt.Errorf("esptool: unexpected char %02X after escape character", bytes[0])
			}
		}
	}
}
