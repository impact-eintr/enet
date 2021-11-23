package main

import (
	"fmt"
	"net"
	"time"

	"github.com/impact-eintr/enet"
)

func listenHeartBeat() {
	ip := net.ParseIP("172.17.0.2")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	for {
		reqMsg := enet.NewMsgPackage(11, []byte("让我访问!!!"))
		buf := enet.NewDataPack().Encode(reqMsg)
		_, err = conn.Write(buf[:])
		if err != nil {
			fmt.Println(err)
		}

		resp := make([]byte, 1024)
		n, err := conn.Read(resp)
		if err != nil {
			fmt.Println(err)
			continue
		}
		resp = resp[:n]

		respMsg := enet.NewDataPack().Decode(resp)
		fmt.Println(string(respMsg.GetData()))

		time.Sleep(2000 * time.Millisecond)
	}

}

func main() {
	go listenHeartBeat()

	// 每 3 秒发送一个文件定位信息
	ip := net.ParseIP("172.17.0.2")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	for {
		reqMsg := enet.NewMsgPackage(20, []byte("文件定位测试"))
		buf := enet.NewDataPack().Encode(reqMsg)
		_, err = conn.Write(buf[:])
		if err != nil {
			fmt.Println(err)
		}

		// 暂时不要回复
		resp := make([]byte, 1024)
		n, err := conn.Read(resp)
		if err != nil {
			fmt.Println(err)
		}
		resp = resp[:n]

		respMsg := enet.NewDataPack().Decode(resp)
		fmt.Println("文件位于:", string(respMsg.GetData()))

		time.Sleep(5 * time.Second)
	}
}
