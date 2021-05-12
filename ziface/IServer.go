package ziface

// 定义一个服务器接口
type IServer interface {
	// 启动
	Start()
	// 停止
	Stop()
	// 运行
	Run()
	// 路由功能
	AddRouter(msgId uint32, router IRouter)
	// 获取当前Server的链接管理器
	GetConnManager() IConnManager
	//设置该Server的连接创建时Hook函数
	SetOnConnStart(func(IConnection))
	//设置该Server的连接断开时的Hook函数
	SetOnConnStop(func(IConnection))
	//调用连接OnConnStart Hook函数
	CallOnConnStart(conn IConnection)
	//调用连接OnConnStop Hook函数
	CallOnConnStop(conn IConnection)
}
