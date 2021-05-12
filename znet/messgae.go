package znet

type Message struct {
	ID      uint32 // 消息ID
	DataLen uint32 // 消息长度
	Data    []byte // 消息的内容
}

func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		ID:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// 获取消息的ID
func (m *Message) GetMsgID() uint32 {
	return m.ID
}

// 获取消息的长度
func (m *Message) GetMsgLen() uint32 {
	return m.DataLen
}

// 获取消息的内容
func (m *Message) GetMsgData() []byte {
	return m.Data
}

// 设置消息的ID
func (m *Message) DetMsgID(mid uint32) {
	m.ID = mid
}

// 设置消息的长度
func (m *Message) SetMsgLen(mlen uint32) {
	m.DataLen = mlen
}

// 设置消息的内容
func (m *Message) SetMsgData(mdata []byte) {
	m.Data = mdata
}
