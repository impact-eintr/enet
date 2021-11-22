package enet

import (
	"encoding/json"
	"io/ioutil"

	"github.com/impact-eintr/enet/iface"
)

type GlobalObj struct {
	TcpServer iface.IServer // 当前Zinx的全局Server对象
	Host      string        // 当前服务器主机IP
	Port      int           // 当前服务器主机监听端口号
	Name      string        // 当前服务器名称
	Version   string        // 当前版本号

	MaxPacketSize    uint32 // 都需数据包的最大值
	MaxConn          int    // 当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 // 业务工作Worker池的数量
	MaxWorkerTaskLen uint32 // 业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    uint32 // 消息缓冲管道长度
	/*
		config file path
	*/
	ConfFilePath string
}

var GlobalObject *GlobalObj

// 读取用户的配置文件
func (g *GlobalObj) Reload() {
	data, err := ioutil.ReadFile("conf/enet.json")
	if err != nil {
		panic(err)
	}
	// 将json数据解析到struct中
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}

/*
	提供init方法，默认加载
*/
func init() {
	// 初始化GlobalObject变量，设置一些默认值
	GlobalObject = &GlobalObj{
		Name:             "enetServerApp",
		Version:          "V1.0",
		Port:             6430,
		Host:             "0.0.0.0",
		MaxConn:          1000,
		MaxPacketSize:    4096,
		WorkerPoolSize:   0, // 默认不开启任务池
		MaxWorkerTaskLen: 0,
		MaxMsgChanLen:    64,
	}
	// 从配置文件中加载一些用户配置的参数
	GlobalObject.Reload()
}
