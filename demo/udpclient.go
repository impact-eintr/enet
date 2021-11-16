package main

import (
	"fmt"
	"net"
	"time"
)

func main() {

	fmt.Println("UDP Client Test ... start")
	time.Sleep(1 * time.Second)

	ip := net.ParseIP("127.0.0.1")
	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	conn.Write([]byte("hello"))
	data := make([]byte, 1024)
	n, err := conn.Read(data)
	fmt.Printf("read %s from <%s>\n", data[:n], conn.RemoteAddr())

}
