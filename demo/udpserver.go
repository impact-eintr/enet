package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

var m = make(map[string]time.Time, 0) // 计数器

// 监听心跳的Router
type LHBRouter struct {
	enet.BaseRouter //一定要先基础BaseRouter
}

// 心跳监控
func (this *LHBRouter) Handle(request iface.IRequest) {
	// 先更新
	locker.Lock()
	m[string(request.GetData())] = time.Now()
	locker.Unlock()
}

// 广播心跳的Router
type BHBRouter struct {
	enet.BaseRouter
}

func (this *BHBRouter) Handle(request iface.IRequest) {
	// 访问这个Router的都是API server
	var s string
	locker.RLock()
	for k := range m {
		s += k + " " // A.A.A.A:a B.B.B.B:b
	}
	locker.RUnlock()

	err := request.GetConnection().SendBuffUdpMsg(10,
		[]byte(s), request.GetRemoteAddr())
	if err != nil {
		fmt.Println(err)
	}
}

// broadcast file location
type BFLRouter struct {
	enet.BaseRouter
}

func (this *BFLRouter) Handle(request iface.IRequest) {
	file := string(request.GetData())
	log.Println(file)
	btest()
}

func ListenHeartBeat() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("udp")

	s.AddRouter(10, &LHBRouter{})
	s.AddRouter(11, &BHBRouter{})
	s.AddRouter(20, &BFLRouter{})

	//2 开启服务
	s.Serve()
}

var locker sync.RWMutex

func btest() error {
	conn, err := net.DialUDP("udp", nil,
		&net.UDPAddr{
			IP:   net.IPv4(192, 168, 46, 255),
			Port: 9000,
		}) // 协议, 发送者,接收者
	defer conn.Close()
	if err != nil {
		return err
	}

	for {
		fmt.Print("input: ")
		var data string
		fmt.Scan(&data)
		_, e := conn.Write([]byte(data))
		if e != nil {
			fmt.Println(e.Error())
			continue
		}
		fmt.Println("发送成功")
	}
}

func main() {
	go ListenHeartBeat()

	for {
		locker.Lock()
		for k, t := range m {
			if t.Add(2 * time.Second).Before(time.Now()) {
				delete(m, k)
				log.Printf("<%s>失效\n", k)
			}
		}
		locker.Unlock()
		time.Sleep(2 * time.Second)
	}

}
