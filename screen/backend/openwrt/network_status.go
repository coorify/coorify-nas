package openwrt

import (
	"encoding/json"
	"io"
	"net/http"
)

type networkDeviceListReply struct {
	Success int
	Result  struct {
		Devices []struct {
			Ipv4addr string
			Mac      string
		}
	}
}

func (w *openwrt) networkDeviceList() (*networkDeviceListReply, error) {
	req, err := http.NewRequest(http.MethodGet, w.combineUrl("/network/device/list"), nil)
	if err != nil {
		return nil, err
	}

	rep, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(rep.Body)
	if err := rep.Body.Close(); err != nil {
		return nil, err
	}

	reply := networkDeviceListReply{}
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}

type networkStatisticsReply struct {
	Success int
	Result  struct {
		Items []struct {
			DownloadSpeed int
			EndTime       int
			StartTime     int
			UploadSpeed   int
		}
	}
}

func (w *openwrt) networkStatistics() (*networkStatisticsReply, error) {
	req, err := http.NewRequest(http.MethodGet, w.combineUrl("/u/network/statistics"), nil)
	if err != nil {
		return nil, err
	}

	rep, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(rep.Body)
	if err := rep.Body.Close(); err != nil {
		return nil, err
	}

	reply := networkStatisticsReply{}
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, err
	}

	return &reply, nil
}

func (w *openwrt) NetworkStatus() (*NetworkStatus, error) {
	s := &NetworkStatus{}

	dev, err := w.networkDeviceList()
	if err != nil {
		return nil, err
	}
	s.Num = uint16(len(dev.Result.Devices))

	sta, err := w.networkStatistics()
	if err != nil {
		return nil, err
	}

	st := sta.Result.Items[len(sta.Result.Items)-1]
	s.Up = n2u(st.UploadSpeed)
	s.Down = n2u(st.DownloadSpeed)

	return s, nil
}

func n2u(v int) uint16 {
	u := uint16(0)
	for v > 1024 {
		u += 0x4000
		v /= 1024
	}

	return uint16(u | uint16(v))
}
