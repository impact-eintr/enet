package main

import (
	"fmt"
	"net"
	"time"

	"github.com/impact-eintr/enet"
)

func listenHeartBeat() {
	ip := net.ParseIP("127.0.0.1")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	for {
		conn, err := net.DialUDP("udp", srcAddr, dstAddr)
		if err != nil {
			fmt.Println(err)
		}

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
		}
		resp = resp[:n]

		respMsg := enet.NewDataPack().Decode(resp)
		fmt.Println(string(respMsg.GetData()), respMsg.GetMsgId())
		conn.Close()

		time.Sleep(2000 * time.Millisecond)
	}

}

func main() {
	go listenHeartBeat()

	// 每 3 秒发送一个文件定位信息
	for {
		ip := net.ParseIP("127.0.0.1")

		srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
		dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

		conn, err := net.DialUDP("udp", srcAddr, dstAddr)
		if err != nil {
			fmt.Println(err)
		}

		reqMsg := enet.NewMsgPackage(20, []byte("文件测试"))
		buf := enet.NewDataPack().Encode(reqMsg)
		_, err = conn.Write(buf[:])
		if err != nil {
			fmt.Println(err)
		}

		// 暂时不要回复
		//resp := make([]byte, 1024)
		//n, err := conn.Read(resp)
		//if err != nil {
		//	fmt.Println(err)
		//}
		//resp = resp[:n]

		//respMsg := enet.NewDataPack().Decode(resp)
		//fmt.Println(string(respMsg.GetData()), respMsg.GetMsgId())
		conn.Close()

		time.Sleep(3 * time.Second)
		fmt.Println("==========================")
	}
}
