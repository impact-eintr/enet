package enet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/impact-eintr/enet/iface"
)

// 封包拆包类实例，暂时不需要成员
type DataPack struct{}

// 获取封包拆包实例
func NewDataPack() *DataPack {
	return &DataPack{}
}

// 获取包头长度
func (dp *DataPack) GetHeadLen() uint32 {
	// Id uint32(4字节) +  DataLen uint32(4字节)
	return 8
}

// 封包方法(压缩数据)
func (dp *DataPack) Pack(msg iface.IMessage) ([]byte, error) {
	//创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	//写msgID
	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	//写dataLen
	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	//写data数据
	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

// 拆包方法(解压数据)
func (dp *DataPack) Unpack(binaryData []byte) (iface.IMessage, error) {
	//创建一个从输入二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	//只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	//读msgID
	if err := binary.Read(dataBuff, binary.BigEndian, &msg.Id); err != nil {
		return nil, err
	}

	//读dataLen
	if err := binary.Read(dataBuff, binary.BigEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	//判断dataLen的长度是否超出我们允许的最大包长度
	if GlobalObject.MaxPacketSize > 0 && msg.DataLen > GlobalObject.MaxPacketSize {
		return nil, errors.New("Too large msg data recieved")
	}

	//这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}

func (dp *DataPack) Decode(buf []byte) iface.IMessage {
	// 解码 构建消息
	msgId := binary.BigEndian.Uint32(buf[0:4])
	msg := &Message{Data: buf[4:], Id: msgId}
	return msg
}

func (dp *DataPack) Encode(msg iface.IMessage) []byte {
	buf := make([]byte, msg.GetDataLen()+4)
	binary.BigEndian.PutUint32(buf[0:4], msg.GetMsgId())
	copy(buf[4:], msg.GetData())
	return buf
}
