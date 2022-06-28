package Landis

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func GetIP_Broadcast() net.IP {
	mask := net.CIDRMask(20, 32)
	s := strings.Split(GetIP_Local().String(), ".")
	i1, _ := strconv.Atoi(s[0])
	i2, _ := strconv.Atoi(s[1])
	i3, _ := strconv.Atoi(s[2])
	i4, _ := strconv.Atoi(s[3])
	ip := net.IP([]byte{byte(i1), byte(i2), byte(i3), byte(i4)})

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

func GetIP_Local() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return net.ParseIP("127.0.0.1")
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().String()
	idx := strings.LastIndex(localAddr, ":")
	return net.ParseIP(localAddr[0:idx])
}
