package util

import (
	"fmt"
	"github.com/zohu/reg"
	"io"
	"net/http"
	"time"
)

func NetIp() string {
	hosts := []string{
		"myexternalip.com/raw",
		"checkip.amazonaws.com",
		"api.ipify.org",
		"ifconfig.me/ip",
		"icanhazip.com",
		"ipinfo.io/ip",
		"ipecho.net/plain",
		"checkipv4.dedyn.io",
	}
	client := http.Client{Timeout: 3 * time.Second}
	for _, host := range hosts {
		resp, err := client.Get(fmt.Sprintf("http://%s", host))
		if err != nil {
			continue
		}
		ip, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if reg.String(string(ip)).IsIpv4().B() {
			return string(ip)
		}
	}
	return ""
}
