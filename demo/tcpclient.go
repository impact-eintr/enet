package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/impact-eintr/enet"
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
		//发封包message消息
		dp := enet.NewDataPack()
		msg, _ := dp.Pack(enet.NewMsgPackage(0, []byte("enet V1.0 Client Test Message")))
		_, err := conn.Write(msg)
		if err != nil {
			fmt.Println("write error err ", err)
			return
		}

		//先读出流中的head部分
		headData := make([]byte, dp.GetHeadLen())
		_, err = io.ReadFull(conn, headData) //ReadFull 会把msg填充满为止
		if err != nil {
			fmt.Println("read head error")
			break
		}
		//将headData字节流 拆包到msg中
		msgHead, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("server unpack err:", err)
			return
		}

		if msgHead.GetDataLen() > 0 {
			//msg 是有data数据的，需要再次读取data数据
			msg := msgHead.(*enet.Message)
			msg.Data = make([]byte, msg.GetDataLen())

			//根据dataLen从io中读取字节流
			_, err := io.ReadFull(conn, msg.Data)
			if err != nil {
				fmt.Println("server unpack data err:", err)
				return
			}

			fmt.Println("==> Recv Msg: ID=", msg.Id, ", len=", msg.DataLen, ", data=", string(msg.Data))
		}

		time.Sleep(1 * time.Second)
	}

}
