package main

import (
	"fmt"
	"log"
	"time"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

//ping test 自定义路由
type PongRouter struct {
	enet.BaseRouter //一定要先基础BaseRouter
}

var m = make(map[string]time.Time, 0) // 计数器

// 心跳监控
func (this *PongRouter) Handle(request iface.IRequest) {
	// 先更新
	m[string(request.GetData())] = time.Now()

	fmt.Printf("来自<%s>的心跳 %s\n", request.GetRemoteAddr().String(), string(request.GetData()))
}

func (this *PongRouter) PostHandle(request iface.IRequest) {
	// 然后检查失效节点
	for k, t := range m {
		fmt.Println(t)
		if t.Add(2 * time.Second).Before(time.Now()) {
			delete(m, k)
			log.Printf("<%s>失效\n", k)
		}
	}
}

func main() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("udp")

	s.AddRouter(0, &PongRouter{})
	s.AddRouter(1, &PongRouter{})

	//2 开启服务
	s.Serve()
}
