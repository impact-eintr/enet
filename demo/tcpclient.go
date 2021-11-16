package main

import (
	"fmt"
	"net"
	"time"
)

func main() {

	fmt.Println("TCP Client Test ... start")
	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp4", "127.0.0.1:6430")
	if err != nil {
		fmt.Println("client start err, exit!")
		return
	}

	for {
		_, err := conn.Write([]byte("hahaha"))
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}

		buf := make([]byte, 512)
		cnt, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read buf error ")
			return
		}

		fmt.Printf(" server call back : %s, cnt = %d\n", buf, cnt)

		time.Sleep(1 * time.Second)
	}
}
