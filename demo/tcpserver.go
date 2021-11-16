package main

import "github.com/impact-eintr/Zinx/enet"

func main() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("[enetv1.0]", "tcp4")

	//2 开启服务
	s.Serve()
}
