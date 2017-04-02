package fnet

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/viphxin/xingo/logger"
)

type PkgData struct {
	Len   uint32
	MsgId uint32
	Data  []byte
}

type PBDataPack struct{}

func NewPBDataPack() *PBDataPack {
	return &PBDataPack{}
}

func (this *PBDataPack) GetHeadLen() int32 {
	return 8
}

func (this *PBDataPack) Unpack(headdata []byte) (interface{}, error) {
	headbuf := bytes.NewReader(headdata)

	head := &PkgData{}

	// 读取Len
	if err := binary.Read(headbuf, binary.LittleEndian, &head.Len); err != nil {
		return nil, err
	}

	// 读取MsgId
	if err := binary.Read(headbuf, binary.LittleEndian, &head.MsgId); err != nil {
		return nil, err
	}

	// 封包太大
	if head.Len > MaxPacketSize {
		return nil, packageTooBig
	}

	return head, nil
}

func (this *PBDataPack) Pack(msgId uint32, data interface{}) (out []byte, err error) {
	outbuff := bytes.NewBuffer([]byte{})
	// 进行编码
	dataBytes := []byte{}
	if data != nil {
		dataBytes, err = proto.Marshal(data.(proto.Message))
	}

	if err != nil {
		logger.Error(fmt.Sprintf("marshaling error:  %s", err))
	}
	// 写Len
	if err = binary.Write(outbuff, binary.LittleEndian, uint32(len(dataBytes))); err != nil {
		return
	}
	// 写MsgId
	if err = binary.Write(outbuff, binary.LittleEndian, msgId); err != nil {
		return
	}

	//all pkg data
	if err = binary.Write(outbuff, binary.LittleEndian, dataBytes); err != nil {
		return
	}

	out = outbuff.Bytes()
	return

}
