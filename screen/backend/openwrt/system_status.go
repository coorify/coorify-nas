package openwrt

import (
	"encoding/json"
	"io"
	"net/http"
)

type systemStatusReply struct {
	Success int
	Result  struct {
		CpuTemperature         int
		CpuUsage               int
		MemAvailablePercentage int
	}
}

func (w *openwrt) SystemStatus() (*SystemStatus, error) {
	req, err := http.NewRequest(http.MethodGet, w.combineUrl("/system/status"), nil)
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

	reply := systemStatusReply{}
	if err := json.Unmarshal(body, &reply); err != nil {
		return nil, err
	}

	if reply.Success != 0 {
		w.Sigin()
		return w.SystemStatus()
	}

	return &SystemStatus{
		Cpu: uint16(reply.Result.CpuUsage),
		Mem: uint16(100 - reply.Result.MemAvailablePercentage),
		Tmp: uint16(reply.Result.CpuTemperature),
	}, nil
}
