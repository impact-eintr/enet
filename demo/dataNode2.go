//package main
//
//import (
//	"fmt"
//	"net"
//	"time"
//
//	"github.com/impact-eintr/enet"
//)
//
//func main() {
//	localhost := "10.29.1.3:12345"
//	ip := net.ParseIP("127.0.0.1")
//
//	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
//	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}
//	for {
//		conn, err := net.DialUDP("udp", srcAddr, dstAddr)
//		if err != nil {
//			fmt.Println(err)
//		}
//
//		msg := enet.NewMsgPackage(10, []byte(localhost))
//		buf := enet.NewDataPack().Encode(msg)
//		_, err = conn.Write(buf[:])
//		if err != nil {
//			fmt.Println(err)
//		}
//		conn.Close()
//
//		time.Sleep(50 * time.Millisecond)
//	}
//}
package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/impact-eintr/enet"
)

func SendHeartBeat() {
	localhost := "10.29.1.3:12345"
	ip := net.ParseIP("127.0.0.1")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	for {
		conn, err := net.DialUDP("udp", srcAddr, dstAddr)
		if err != nil {
			fmt.Println(err)
		}

		msg := enet.NewMsgPackage(10, []byte(localhost))
		buf := enet.NewDataPack().Encode(msg)
		_, err = conn.Write(buf[:])
		if err != nil {
			fmt.Println(err)
		}
		conn.Close()

		time.Sleep(50 * time.Millisecond)
	}

}

func main() {
	go SendHeartBeat()

	// TODO 准备接受广播 每个dataNode 是一个server
	// 解析得到UDP地址
	addr, err := net.ResolveUDPAddr("udp", ":9000")
	if err != nil {
		log.Fatal(err)
	}

	// 在UDP地址上建立UDP监听,得到连接
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	// 建立缓冲区
	buffer := make([]byte, 1024)

	for {
		//从连接中读取内容,丢入缓冲区
		i, udpAddr, e := conn.ReadFromUDP(buffer)
		// 第一个是字节长度,第二个是udp的地址
		if e != nil {
			continue
		}
		fmt.Printf("来自%v,读到的内容是:%s\n", udpAddr, buffer[:i])

		// 向客户端返回消息
		conn.WriteToUDP([]byte("hello"), udpAddr)
	}
}
