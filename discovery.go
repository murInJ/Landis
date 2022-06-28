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
	innerPort      int
	localIP        net.IP
	discoveryMap   map[string]time.Time
	controlChannel chan int
	localAdress    string
	debug          bool
	broadcastIP    net.IP
	outerPort      int
}

func (d *Discovery) SetDebug(val bool) {
	d.debug = val
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

	laddr := net.UDPAddr{
		IP:   d.localIP,
		Port: d.outerPort,
	}

	raddr := net.UDPAddr{
		IP:   d.broadcastIP,
		Port: d.outerPort,
	}

	conn, err := net.DialUDP("udp", &laddr, &raddr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	_, err = conn.Write([]byte(d.localIP.String() + ":" + strconv.Itoa(d.outerPort)))
	if err != nil {
		log.Fatal(err)
	}

	if d.debug {
		fmt.Printf("%s LAN broadcast %s\n",
			color.New(color.FgHiCyan).Sprintf("Landis:"),
			color.New(color.FgYellow).Sprintf(d.localIP.String()+":"+strconv.Itoa(d.outerPort)))
	}
}

func (d *Discovery) recvBoardcast() {
	laddr := net.UDPAddr{
		IP:   d.localIP,
		Port: d.innerPort,
	}
	conn, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case controlInfo := <-d.controlChannel:
			if controlInfo == 0 {
				conn.Close()
				return
			}
		default:
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)

			if err != nil {
				log.Fatal(err)
			}
			address := string(buf[:n])
			if address != d.localAdress {
				d.push(address, time.Second*10)
				if d.debug {
					fmt.Printf("%s recv broadcast %s\n",
						color.New(color.FgHiCyan).Sprintf("Landis:"),
						color.New(color.FgYellow).Sprintf(string(buf[:n])))
				}
			}
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

func NewLanDiscovery(outerPort, innerPort int) *Discovery {
	localIP := GetIP_Local()
	//localIP, _ := GetIp_Public()
	//localIP := net.ParseIP("0.0.0.0")
	return &Discovery{
		innerPort:      innerPort,
		outerPort:      outerPort,
		discoveryMap:   make(map[string]time.Time),
		localAdress:    localIP.String() + ":" + strconv.Itoa(outerPort),
		localIP:        localIP,
		controlChannel: make(chan int),
		debug:          false,
		broadcastIP:    GetIP_Broadcast(),
	}
}
