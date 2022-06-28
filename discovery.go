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
	localIP        net.IP
	discoveryMap   map[string]time.Time
	controlChannel chan int
	localAdress    string
	debug          bool
	broadcastIP    net.IP
}

func (d *Discovery) SetDebug(val bool) {
	d.debug = val
}

func (d *Discovery) update() {
	newMap := make(map[string]time.Time)
	for k := range d.discoveryMap {
		if d.discoveryMap[k].After(time.Now()) {
			newMap[k] = d.discoveryMap[k]
		}
	}
	d.discoveryMap = nil
	d.discoveryMap = newMap
}

func (d *Discovery) push(address string, duration time.Duration) {
	timeout := time.Now().Add(duration)
	d.discoveryMap[address] = timeout
}

func (d *Discovery) List() []string {
	d.update()
	var list []string
	for k := range d.discoveryMap {
		list = append(list, k)
	}
	return list
}

func (d *Discovery) boardcast() {
	socket, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   d.broadcastIP,
		Port: d.port,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer socket.Close()
	str := GetIP_Local().String() + ":" + strconv.Itoa(d.port)
	sendData := []byte(str)
	_, err = socket.Write(sendData) // 发送数据
	if err != nil {
		log.Fatal(err)
	}

	if d.debug {
		fmt.Printf("%s LAN broadcast %s\n",
			color.New(color.FgHiCyan).Sprintf("Landis:"),
			color.New(color.FgYellow).Sprintf(str))
	}
}

func (d *Discovery) recvBoardcast() {
	listen, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: d.port,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case controlInfo := <-d.controlChannel:
			if controlInfo == 0 {
				listen.Close()
				return
			}
		default:
			var data [1024]byte
			n, _, err := listen.ReadFromUDP(data[:]) // 接收数据
			if err != nil {
				log.Fatal(err)
			}
			address := string(data[:n])
			if address != d.localAdress {
				d.push(address, time.Second*10)
				if d.debug {
					fmt.Printf("%s recv broadcast %s\n",
						color.New(color.FgHiCyan).Sprintf("Landis:"),
						color.New(color.FgYellow).Sprintf(address))
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

func NewLanDiscovery(port int) *Discovery {
	return &Discovery{
		port: port,

		discoveryMap:   make(map[string]time.Time),
		localAdress:    GetIP_Local().String() + ":" + strconv.Itoa(port),
		localIP:        GetIP_Local(),
		controlChannel: make(chan int),
		debug:          false,
		broadcastIP:    GetIP_Broadcast(),
	}
}
