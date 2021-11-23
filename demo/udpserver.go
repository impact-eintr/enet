package main

import (
	"fmt"
	"log"
	"net"
	"strings"
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

var bflm = make(map[string]chan []byte)
var bflLocker sync.RWMutex

func (this *BFLRouter) Handle(request iface.IRequest) {
	btest(request.GetData()) // 发一个定位广播

	bflm[string(request.GetData())] = make(chan []byte)

	select {
	case res := <-bflm[string(request.GetData())]:
		// 把定位结果返回
		request.GetConnection().SendBuffUdpMsg(20,
			res, request.GetRemoteAddr())
		delete(bflm, string(request.GetData()))
	case <-time.Tick(1 * time.Second):
		// 这里可以倒计时:
		request.GetConnection().SendBuffUdpMsg(20,
			[]byte("没有找到"), request.GetRemoteAddr())
		delete(bflm, string(request.GetData()))
	}

}

func btest(data []byte) error {
	conn, err := net.DialUDP("udp", nil,
		&net.UDPAddr{
			IP:   net.IPv4(172, 17, 255, 255),
			Port: 9000,
		}) // 协议, 发送者,接收者
	defer conn.Close()
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	if err != nil {
		return err
	}
	fmt.Println("广播发送成功")

	return nil
}

// response file location
type RFLRouter struct {
	enet.BaseRouter
}

func (this *RFLRouter) Handle(request iface.IRequest) {
	rtest(request.GetData())
}

// RFL 21
func rtest(data []byte) {
	s := strings.Split(string(data), "\n")
	select {
	case bflm[s[0]] <- []byte(s[1]):
	case <-time.Tick(1 * time.Second):
	}
	//bflLocker.Lock()
	//if _, ok := bflm[s[0]]; ok {
	//	bflm[s[0]] <- data
	//}
	//bflLocker.Unlock()
}

func ListenHeartBeat() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("udp")

	s.AddRouter(10, &LHBRouter{})
	s.AddRouter(11, &BHBRouter{})
	s.AddRouter(20, &BFLRouter{})
	s.AddRouter(21, &RFLRouter{})

	//2 开启服务
	s.Serve()
}

var locker sync.RWMutex

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
