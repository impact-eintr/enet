package ziface

type IDataPack interface {
	// 获取数据包头长度
	GetHeadLen() uint32

	// 封装数据包
	Pack(IMessage) ([]byte, error)

	// 拆解数据包
	Unpack([]byte) (IMessage, error)
}
