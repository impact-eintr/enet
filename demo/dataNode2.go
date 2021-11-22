package main

import (
	"fmt"
	"net"
	"time"

	"github.com/impact-eintr/enet"
)

func main() {
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
