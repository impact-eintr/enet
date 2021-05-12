package znet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/impact-eintr/Zinx/utils"
	"github.com/impact-eintr/Zinx/ziface"
)

type DataPack struct{}

// 工厂方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

func (d *DataPack) GetHeadLen() uint32 {
	// Datalen uint32(4bytes) + ID uint32(4 bytes)
	return 8
}

func (d *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	// 创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	// 将datalen写入databuff
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgLen()); err != nil {
		return nil, err
	}

	// 将MsgID写入databuff
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgID()); err != nil {
		return nil, err
	}

	// 将data写入databuff
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil

}

// 拆包方法 将包头信息读出 再根据head中的data长度 读取data
func (d *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	// 创建一个从输入读取二进制数据的ioReader
	dataBuff := bytes.NewReader(binaryData)

	// 解压head信息
	msg := &Message{}

	// 读取datalen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 读取msgid
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.ID); err != nil {
		return nil, err
	}

	if utils.GlobalConf.MaxPackageSize > 0 && msg.DataLen > utils.GlobalConf.MaxPackageSize {
		return nil, errors.New("too long msg data recv")
	}

	return msg, nil
}
