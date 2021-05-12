package ziface

type IRequest interface {
	// 获取当前连接
	GetConnection() IConnection
	// 获取请求的消息数据
	GetMsgData() []byte
	// 获取请求的消息ID
	GetMsgID() uint32
}
