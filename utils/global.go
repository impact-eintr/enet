package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/impact-eintr/Zinx/ziface"
)

type Global struct {
	/*
		Zinx
	*/
	Version        string `json:"version"`          // 当前Zinx的版本号
	MaxConn        int    `json:"max_conn"`         // 当前服务器允许的最大连接数
	MaxPackageSize uint32 `json:"max_package_size"` // 当前Zinx框架数据包的最大值
	WorkerPoolSize uint32 `json:"worker_pool_size"` //当前Zinx资源池限制
	TaskQueueSize  uint32 `json:"task_queue_size"`  //当前Zinx等待队列限制

	/*
		Server
	*/
	TcpServer ziface.IServer // 当前zinx全局的Server对象
	Host      string         `json:"host"` // 当前服务器监听的IP
	Port      int            `json:"port"` // 当前服务器监听的Port
	Name      string         `json:"name"` // 当前服务器的名称

}

var GlobalConf *Global

func (g *Global) Reload() {
	data, err := ioutil.ReadFile("conf/zinx.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, &GlobalConf)
	if err != nil {
		panic(err)
	}

}

func init() {
	fmt.Println("test")
	GlobalConf = &Global{
		Name:           "ZinxApp",
		MaxConn:        1000,
		MaxPackageSize: 4096,
		WorkerPoolSize: 8,
		TaskQueueSize:  1024,
		Version:        "1.0",
		Host:           "0.0.0.0",
		Port:           8889,
	}

	// 尝试从conf/zinx.json中加载一些用户自定义的配置
	GlobalConf.Reload()
}
