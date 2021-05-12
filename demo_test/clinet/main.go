package main

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/impact-eintr/Zinx/znet"
)

func main() {
	// 直接连接远程服务器 得到一个conn连接
	conn, err := net.Dial("tcp", "127.0.0.1:8999")
	if err != nil {
		fmt.Println("write conn err:", err)
		return
	}

	for {
		dp := znet.NewDataPack()

		msg, err := dp.Pack(znet.NewMsgPackage(0, []byte(conn.LocalAddr().String())))
		if err != nil {
			fmt.Println("write conn err", err)
			return
		}

		if _, err := conn.Write(msg); err != nil {
			fmt.Println("write conn err", err)
			return
		}

		// 先读取流中的head部分 得到ID dataLen
		binaryHead := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(conn, binaryHead); err != nil {
			fmt.Println("read head error", err)
			return
		}

		// 将二进制的head拆包到msg结构体中
		msgHead, err := dp.Unpack(binaryHead)
		if err != nil {
			fmt.Println("client Unpack msgHead error", err)
			break
		}

		// 再根据dataLen进行第二次读取 将data取出来
		if msgHead.GetMsgLen() > 0 {
			msg := msgHead.(*znet.Message)
			msg.Data = make([]byte, msg.GetMsgLen())

			if _, err := io.ReadFull(conn, msg.Data); err != nil {
				fmt.Println("read head error", err)
				return
			}

			fmt.Printf("Recv Server Msg: ID:%v len:%v data:%v\n",
				msg.ID, msg.DataLen, string(msg.Data))
		}

		time.Sleep(time.Second)
	}

}
