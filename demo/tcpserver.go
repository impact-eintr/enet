package main

import (
	"fmt"
	"log"
	"time"

	"github.com/impact-eintr/enet"
)

//ping test 自定义路由
type PingRouter struct {
	*enet.BaseRouter //一定要先基础BaseRouter
}

//Test Handle
func (this *PingRouter) Handle(request enet.IRequest) {
	fmt.Println("Call PingRouter Handle 🥵")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))

	//回写数据
	err := request.GetConnection().SendBuffMsg(1, []byte("ping...ping...ping"))
	if err != nil {
		fmt.Println(err)
	}
}

type HelloRouter struct {
	*enet.BaseRouter
}

func (this *HelloRouter) Handle(request enet.IRequest) {
	fmt.Println("Call HelloRouter Handle")
	//先读取客户端的数据，再回写ping...ping...ping
	fmt.Println("recv from client : msgId=", request.GetMsgID(), ", data=", string(request.GetData()))
	defer log.Println("Goroutine Exiting...")

	// TODO 如果这里面是一个阻塞任务 怎么办
	for {
		select {
		case <-this.Exit(request.GetConnection().GetConnID()):
			return
		default:
			err := request.GetConnection().SendBuffMsg(1, []byte("Hello Router V1.0"))
			if err != nil {
				fmt.Println(err)
				return
			}
			time.Sleep(time.Second)
		}
	}
}

//创建连接的时候执行
//func DoConnectionBegin(conn enet.IConnection) {
//	fmt.Println("DoConnecionBegin is Called ... ")
//
//	fmt.Println("Set conn Name, Home done!")
//	conn.SetProperty("Name", "Impact-EINTR")
//
//	err := conn.SendBuffMsg(2, []byte("DoConnection BEGIN..."))
//	if err != nil {
//		fmt.Println(err)
//	}
//}

//连接断开的时候执行
//func DoConnectionLost(conn enet.IConnection) {
//	//============在连接销毁之前，查询conn的Name，Home属性=====
//	if name, err := conn.GetProperty("Name"); err == nil {
//		fmt.Println("Conn Property Name = ", name)
//	}
//
//	fmt.Println("DoConneciotnLost is Called ... ")
//}

func main() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("tcp4")

	//注册链接hook回调函数
	//s.SetOnConnStart(DoConnectionBegin)
	//s.SetOnConnStop(DoConnectionLost)

	// 添加路由
	s.AddRouter(0, &PingRouter{enet.NewBaseRouter()})
	s.AddRouter(1, &HelloRouter{enet.NewBaseRouter()})

	//2 开启服务
	s.Serve()
}
