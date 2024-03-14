package openwrt

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

type versionReply struct {
	Success int
	Result  struct {
		FirmwareVersion string
		KernelVersion   string
		Model           string
	}
}

func (w *openwrt) Sigin() error {
	payload := url.Values{}
	payload.Set("luci_username", w.option.Username)
	payload.Set("luci_password", w.option.Password)

	reader := strings.NewReader(payload.Encode())
	req, err := http.NewRequest(http.MethodPost, w.baseUrl(), reader)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	rep, err := w.client.Do(req)
	if err != nil {
		return err
	}

	if err := rep.Body.Close(); err != nil {
		return err
	}

	req, err = http.NewRequest(http.MethodGet, w.combineUrl("/u/system/version"), nil)
	if err != nil {
		return err
	}

	rep, err = w.client.Do(req)
	if err != nil {
		return err
	}

	defer rep.Body.Close()
	body, _ := io.ReadAll(rep.Body)
	reply := versionReply{}
	if err := json.Unmarshal(body, &reply); err != nil {
		return err
	}

	if reply.Success != 0 {
		return fmt.Errorf("openwrt: sigin fail")
	}

	logrus.Infof("openwrt: firmware(%s)", reply.Result.FirmwareVersion)
	logrus.Infof("openwrt: kernel(%s)", reply.Result.KernelVersion)
	logrus.Infof("openwrt: model(%s)", reply.Result.Model)

	return nil
}
