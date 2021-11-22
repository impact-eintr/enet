package main

import (
	"fmt"
	"net"
	"time"

	"github.com/impact-eintr/enet"
)

func main() {
	//ip := net.ParseIP("192.168.46.255")
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

		time.Sleep(50 * time.Millisecond)
	}
}
