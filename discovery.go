package Landis

import (
	"fmt"
	"github.com/fatih/color"
	"log"
	"net"
	"strconv"
	"time"
)

type LanDiscovery interface {
	push(address string, duration time.Duration)
	List() []string
	boardcast()
	recvBoardcast()
	Start()
	Close()
}

type Discovery struct {
	port           int
	broadcastIP    net.IP
	publicIP       net.IP
	discoveryMap   map[string]time.Time
	isOnline       bool
	controlChannel chan int
}

func (d *Discovery) push(address string, duration time.Duration) {
	timeout := time.Now().Add(duration)
	d.discoveryMap[address] = timeout

	newMap := make(map[string]time.Time)
	for k := range d.discoveryMap {
		if d.discoveryMap[k].After(time.Now()) {
			newMap[k] = d.discoveryMap[k]
		}
	}
	d.discoveryMap = nil
	d.discoveryMap = newMap
}

func (d *Discovery) List() []string {
	var list []string
	for k := range d.discoveryMap {
		list = append(list, k)
	}
	return list
}

func (d *Discovery) boardcast() {
	var err error
	d.publicIP, err = GetIp_Public()
	if err != nil {
		d.isOnline = false
		d.publicIP = net.ParseIP("127.0.0.1")
	}

	d.broadcastIP = GetIP_Broadcast()

	laddr := net.UDPAddr{
		IP:   d.publicIP,
		Port: d.port,
	}

	raddr := net.UDPAddr{
		IP:   d.broadcastIP,
		Port: d.port,
	}

	conn, err := net.DialUDP("udp", &laddr, &raddr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	_, err = conn.Write([]byte(d.publicIP.String() + ":" + strconv.Itoa(d.port)))
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Discovery) recvBoardcast() {
	laddr := net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: d.port,
	}
	conn, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		log.Fatal(err)
	}
	buf := make([]byte, 1024)
	for {
		select {
		case controlInfo := <-d.controlChannel:
			if controlInfo == 0 {
				conn.Close()
				return
			}
		default:
			n, err := conn.Read(buf)
			if err != nil {
				log.Fatal(err)
			}
			d.discoveryMap[string(buf[:n])] = time.Now().Add(time.Duration(time.Second * 10))
		}
	}
}

func (d *Discovery) Start() {
	go d.recvBoardcast()

	go func() {
		for {
			select {
			case controlInfo := <-d.controlChannel:
				if controlInfo == 0 {
					return
				}
			default:
				d.boardcast()
				time.Sleep(time.Duration(time.Second * 5))
			}

		}
	}()

	fmt.Printf("%s LAN Discovery service start\n",
		color.New(color.FgHiCyan).Sprintf("Landis:"))
}

func (d *Discovery) Close() {
	d.controlChannel <- 0
	fmt.Printf("%s LAN Discovery service close\n",
		color.New(color.FgHiCyan).Sprintf("Landis:"))
}

func NewLanDiscovery(port int) *Discovery {
	return &Discovery{
		port: port,
	}
}
