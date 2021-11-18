package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	localhost := "10.29.1.3:12345"
	ip := net.ParseIP("192.168.47.255")

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

		time.Sleep(time.Second)

		ch := make(chan struct{})

		go func() {
			resp := make([]byte, 1024)
			_, _, err = conn.ReadFromUDP(resp)
			if err != nil {
				fmt.Println(err)
			}
			log.Println(string(resp[4:]))
			ch <- struct{}{}
		}()

		//使用time.After：
		select {
		case <-ch:
		case <-time.After(1 * time.Second):
			fmt.Println("timed out")
		}

	}
}
