package openwrt

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"

	"github.com/coorify/be/option"
)

type SystemStatus struct {
	Cpu uint16
	Mem uint16
	Tmp uint16
}

type NetworkStatus struct {
	Num  uint16
	Up   uint16
	Down uint16
}

type Client interface {
	Sigin() error

	SystemStatus() (*SystemStatus, error)
	NetworkStatus() (*NetworkStatus, error)
}

type openwrt struct {
	option  *option.OpenWrtOption
	cookies *cookiejar.Jar
	client  *http.Client
}

func NewClient(o *option.OpenWrtOption) Client {
	cookies, _ := cookiejar.New(nil)

	return &openwrt{
		option:  o,
		cookies: cookies,
		client: &http.Client{
			Jar: cookies,
		},
	}
}

func (w *openwrt) baseUrl() string {
	return fmt.Sprintf("%s/cgi-bin/luci/", w.option.Host)
}

func (w *openwrt) combineUrl(url string) string {
	return fmt.Sprintf("%s%s%s", w.baseUrl(), w.option.OS, url)
}
