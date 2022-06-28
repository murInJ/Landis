package Landis

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
)

func GetIP_Broadcast() net.IP {
	mask := net.CIDRMask(20, 32)
	ip := net.IP([]byte{140, 45, 32, 0})

	broadcast := net.IP(make([]byte, 4))
	for i := range ip {
		broadcast[i] = ip[i] | ^mask[i]
	}

	return broadcast
}

func GetIp_Public() (net.IP, error) {
	responseClient, errClient := http.Get("http://ip.dhcp.cn/?ip") // 获取外网 IP
	if errClient != nil {
		return nil, errClient
	}
	// 程序在使用完 response 后必须关闭 response 的主体。
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(responseClient.Body)

	body, _ := ioutil.ReadAll(responseClient.Body)
	clientIP := fmt.Sprintf("%s", string(body))
	return net.ParseIP(clientIP), nil
}
