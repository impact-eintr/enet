# enet
enet 跟着 zinx 写的 添加了UDP通信

## 示例

> 下面的示例是一个心跳监听+资源定位的demo

``` tcsh
生产者1
	   \
	     <- 广播消息- -发送心跳/资源定位-> udpServer <-定位资源- -检查有多少节点-> 消费者
	   /
生产者2


```


### 消息服务器
``` go
package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/impact-eintr/enet"
	"github.com/impact-eintr/enet/iface"
)

var m = make(map[string]time.Time, 0) // 计数器

// 监听心跳的Router
type LHBRouter struct {
	enet.BaseRouter //一定要先基础BaseRouter
}

// 心跳监控
func (this *LHBRouter) Handle(request iface.IRequest) {
	// 先更新
	locker.Lock()
	m[string(request.GetData())] = time.Now()
	locker.Unlock()
}

// 广播心跳的Router
type BHBRouter struct {
	enet.BaseRouter
}

func (this *BHBRouter) Handle(request iface.IRequest) {
	// 访问这个Router的都是API server
	var s string
	locker.RLock()
	for k := range m {
		s += k + " " // A.A.A.A:a B.B.B.B:b
	}
	locker.RUnlock()

	err := request.GetConnection().SendBuffUdpMsg(10,
		[]byte(s), request.GetRemoteAddr())
	if err != nil {
		fmt.Println(err)
	}
}

// broadcast file location
type BFLRouter struct {
	enet.BaseRouter
}

var bflm = make(map[string]chan []byte)
var bflLocker sync.RWMutex

func (this *BFLRouter) Handle(request iface.IRequest) {
	btest(request.GetData()) // 发一个定位广播

	bflm[string(request.GetData())] = make(chan []byte)

	select {
	case res := <-bflm[string(request.GetData())]:
		// 把定位结果返回
		request.GetConnection().SendBuffUdpMsg(20,
			res, request.GetRemoteAddr())
		delete(bflm, string(request.GetData()))
	case <-time.Tick(1 * time.Second):
		// 这里可以倒计时:
		request.GetConnection().SendBuffUdpMsg(20,
			[]byte("没有找到"), request.GetRemoteAddr())
		delete(bflm, string(request.GetData()))
	}

}

func btest(data []byte) error {
	conn, err := net.DialUDP("udp", nil,
		&net.UDPAddr{
			IP:   net.IPv4(172, 17, 255, 255),
			Port: 9000,
		}) // 协议, 发送者,接收者
	defer conn.Close()
	if err != nil {
		return err
	}

	_, err = conn.Write(data)
	if err != nil {
		return err
	}
	fmt.Println("广播发送成功")

	return nil
}

// response file location
type RFLRouter struct {
	enet.BaseRouter
}

func (this *RFLRouter) Handle(request iface.IRequest) {
	rtest(request.GetData())
}

// RFL 21
func rtest(data []byte) {
	s := strings.Split(string(data), "\n")
	select {
	case bflm[s[0]] <- []byte(s[1]):
	case <-time.Tick(1 * time.Second):
	}
	//bflLocker.Lock()
	//if _, ok := bflm[s[0]]; ok {
	//	bflm[s[0]] <- data
	//}
	//bflLocker.Unlock()
}

func ListenHeartBeat() {
	//1 创建一个server 句柄 s
	s := enet.NewServer("udp")

	s.AddRouter(10, &LHBRouter{})
	s.AddRouter(11, &BHBRouter{})
	s.AddRouter(20, &BFLRouter{})
	s.AddRouter(21, &RFLRouter{})

	//2 开启服务
	s.Serve()
}

var locker sync.RWMutex

func main() {
	go ListenHeartBeat()

	for {
		locker.Lock()
		for k, t := range m {
			if t.Add(2 * time.Second).Before(time.Now()) {
				delete(m, k)
				log.Printf("<%s>失效\n", k)
			}
		}
		locker.Unlock()
		time.Sleep(2 * time.Second)
	}

}

```

### 消费端

``` go
package main

import (
	"fmt"
	"net"
	"time"

	"github.com/impact-eintr/enet"
)

func listenHeartBeat() {
	ip := net.ParseIP("172.17.0.2")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	for {
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
			continue
		}
		resp = resp[:n]

		respMsg := enet.NewDataPack().Decode(resp)
		fmt.Println(string(respMsg.GetData()))

		time.Sleep(2000 * time.Millisecond)
	}

}

func main() {
	go listenHeartBeat()

	// 每 3 秒发送一个文件定位信息
	ip := net.ParseIP("172.17.0.2")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	for {
		reqMsg := enet.NewMsgPackage(20, []byte("文件定位测试"))
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
		fmt.Println("文件位于:", string(respMsg.GetData()))

		time.Sleep(5 * time.Second)
	}
}

```

### 生产端1(有资源)

``` go
package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/impact-eintr/enet"
)

func SendHeartBeat() {
	localhost := "10.29.1.2:12345"
	ip := net.ParseIP("172.17.0.2")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	for {
		conn, err := net.DialUDP("udp", srcAddr, dstAddr)
		if err != nil {
			fmt.Println(err)
		}

		msg := enet.NewMsgPackage(10, []byte(localhost)) // LBH
		buf := enet.NewDataPack().Encode(msg)
		_, err = conn.Write(buf[:])
		if err != nil {
			fmt.Println(err)
		}
		conn.Close()

		time.Sleep(1000 * time.Millisecond)
	}
}

func SendFileLocation(file []byte) {
	localhost := "10.29.1.2:12345"
	ip := net.ParseIP("172.17.0.2")

	srcAddr := &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	dstAddr := &net.UDPAddr{IP: ip, Port: 6430}

	conn, err := net.DialUDP("udp", srcAddr, dstAddr)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	file = append(file, '\n')
	file = append(file, []byte(localhost)...)
	msg := enet.NewMsgPackage(21, file) // RFL
	buf := enet.NewDataPack().Encode(msg)
	_, err = conn.Write(buf[:])
	if err != nil {
		fmt.Println(err)
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

		// Node1 有这个文件
		SendFileLocation(buffer[:i])
	}
}

```

### 生产端2(无资源)

``` go
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
	ip := net.ParseIP("172.17.0.2")

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

		time.Sleep(1 * time.Second)
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
	}
}

```

## v0.2
### 简单的连接封装和业务绑定
> 连接的模块
- 方法
    - 启动连接
    - 停止连接
    - 获取当前连接的conn对象(socket)
    - 得到连接ID
    - 得到客户端连接的地址和端口
    - 发送数据的方法
- 属性
    - socket TCP套接字
    - 连接的ID
    - 当前连接的状态
    - 与当前连接绑定的处理业务方法
    - 等待连接被动退出的channel

## v0.3

> 基础router模块

- Ruquest请求封装(将链接和数据绑定在一起)
    - 属性
        - 连接IConnection
        - 请求数据
    - 方法
        - 得到当前连接
        - 得到当前数据
- Router模块
    - 抽象的IRouter
        -  处理业务之前的方法
        -  处理业务的方法
        -  处理业务之后的方法
    - 具体的BaseRouter(作为具体实现的基类)
        -  处理业务之前的方法
        -  处理业务的方法
        -  处理业务之后的方法
- zinx集成router模块
    - Iserver增添路由功能
    - Server类增添Router成员
    - Commection类绑定一个Router成员
    - 在Connection调用 已经注册的Router处理业务

## v0.4
> 增添全局配置

## v0.5

> 消息封装

- 定义一个消息的结构
    - 属性
        - 消息的ID
        - 消息长度
        - 消息的内容
    - 方法
        - Setter
        - Getter
- 将消息封装机制集成到Zinx框架中
    - 将Message添加到Request中
    - 修改连接读取数据的机制 将之前的单纯读取byte改为拆包读取方式
    - 连接的发包机制 将发送的消息进行打包 再发送

## v0.6

> 消息管理模块

- 属性
    - 集合-消息ID和对应的router的关系 map
- 方法
    - 根据msgID来索引调度路由方法
    - 添加路由方法到map集合中
> 将消息管理机制集成到Zinx框架中
1. 将server模块中的Router属性 替换成MsgHandler属性
2. 将server之前的AddRouter修改成AddRouter--AddRouter(msgId unit32, router ziface.IRouter)
3. 将connection模块Router属性 替换成MsgHandler 修改Connection方法
4. Connection的之前调度Router的业务替换成MsgHandler调度 修改StartReader方法

## v0.7
> Zinx读写分离

![Zinx读写分离](https://img.kancloud.cn/80/28/8028019d6bfce107ebc1bf5a15fd8940_1024x768.jpeg)

1. 添加一个Reader与Write之间通信的channel
2. 添加一个Writer Goroutine
3. Reader由之前直接发送给客户端 改为发送给通信Channel
4. 启动Reader和Writer一同工作


## v0.8
> 消息队列以及多任务

![消息队列](https://img.kancloud.cn/70/6c/706cb06abebcb8c1b7dd22c23d79cf48_1024x768.jpeg)

1. 创建一个消息队列
- MsgHandler消息管理模块
    - 增加属性
        - 消息队列
        - worker工作池的数量
2. 创建多任务worker的工作池并且启动
- 创建一个Worker的工作池并且启动
    - 根据Workerpoolsize的数量去创建Worker
    - 每一个Worker都开启一个协程负载
        - 阻塞等待与当前Worker对应的channel来消息
        - 一旦有消息到来mworker应该处理当前消息对应的业务
3. 将之前的发送消息，全部都改成发送给消息队列和Worker工作池来处理
- 定义一个方法 将消息发送给消息队列工作池
    - 保证每个worker所受的request任务书均衡(平均分配) 
    - 将消息发送给对应的队列 

> 将消息队列机制集成到Zinx框架中
- 开启并调用消息队列
- 将从客户端接收的数据 发送给当前的Worker工作池来处理

## v0.9
> 创建一个链接管理模块 
- 属性
    - 已经创建的Connection集合
    - 针对map的互斥锁
- 方法
    - 添加链接
    - 删除链接
    - 根据链接ID查找对应的链接
    - 总链接个数
    - 清除全部的链接
> 将链接管理模块集成到Zinx框架中

> 给Zinx框架提供 创建链接之后/销毁链接之前 的钩子函数

>

## v1.0

