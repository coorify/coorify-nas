package device

import (
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"go.bug.st/serial/enumerator"
)

var ErrPortNotFound = errors.New("port not found")

func selectPort() (string, error) {
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil || len(ports) == 0 {
		return "", ErrPortNotFound
	}

	for _, port := range ports {
		if port.IsUSB && port.PID == "1001" {
			return port.Name, nil
		}
	}

	return "", ErrPortNotFound
}

func WaitPort() string {

	port, err := selectPort()
	for err != nil {
		logrus.Warn("device: wait port")
		time.Sleep(5 * time.Second)

		port, err = selectPort()
	}

	logrus.Infof("device: use port(%s)", port)
	return port
}
