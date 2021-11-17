package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func main() {
	localhost := "10.29.1.2:12345"
	ip := net.ParseIP("10.29.255.255")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	for {
		conn, err := net.ListenUDP("udp", srcAddr)
		if err != nil {
			fmt.Println(err)
		}

		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf[0:4], 1)
		buf = append(buf, []byte(localhost)...)
		_, err = conn.WriteToUDP(buf[:], dstAddr)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(1 * time.Second)
	}
}
