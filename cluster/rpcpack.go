package cluster

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/viphxin/xingo/fnet"
	"github.com/viphxin/xingo/iface"
	"encoding/gob"
)

type RpcData struct {
	MsgType RpcSignal              `json:"msgtype"`
	Key     string                 `json:"key,omitempty"`
	Target  string                 `json:"target,omitempty"`
	Args    []interface{}          `json:"args,omitempty"`
	Result  map[string]interface{} `json:"result,omitempty"`
}

type RpcPackege struct {
	Len  int32
	Data []byte
}

type RpcRequest struct {
	Fconn   iface.IWriter
	Rpcdata *RpcData
}

func (this *RpcRequest)GetWriter() iface.IWriter{
	return this.Fconn
}
func (this *RpcRequest)GetMsgType() int32{
	return int32(this.Rpcdata.MsgType)
}
func (this *RpcRequest)GetKey() string{
	return this.Rpcdata.Key
}
func (this *RpcRequest)GetTarget() string{
	return this.Rpcdata.Target
}

func (this *RpcRequest)GetArgs() []interface{}{
	return this.Rpcdata.Args
}

func (this *RpcRequest)GetResult() map[string]interface{}{
	return this.Rpcdata.Result
}

func (this *RpcRequest)PushReturn(key string, value interface{}){
	if this.Rpcdata.Result == nil{
		this.Rpcdata.Result = make(map[string]interface{}, 3)
	}
	this.Rpcdata.Result[key] = value
}

type RpcDataPack struct{}

func NewRpcDataPack() *RpcDataPack {
	return &RpcDataPack{}
}

func (this *RpcDataPack) GetHeadLen() int32 {
	return 4
}

func (this *RpcDataPack) Unpack(headdata []byte) (interface{}, error) {
	headbuf := bytes.NewReader(headdata)

	rp := &RpcPackege{}

	// 读取Len
	if err := binary.Read(headbuf, binary.LittleEndian, &rp.Len); err != nil {
		return nil, err
	}

	// 封包太大
	if rp.Len > fnet.MaxPacketSize {
		return nil, errors.New("rpc packege too big!!!")
	}

	return rp, nil
}

//func (this *RpcDataPack) Pack(msgId uint32, pkg interface{}) (out []byte, err error) {
//	outbuff := bytes.NewBuffer([]byte{})
//	// 进行编码
//	dataBytes := []byte{}
//	data := pkg.(*RpcData)
//	if data != nil {
//		dataBytes, err = json.Marshal(data)
//	}
//
//	if err != nil {
//		fmt.Println(fmt.Sprintf("json marshaling error:  %s", err))
//	}
//	// 写Len
//	if err = binary.Write(outbuff, binary.LittleEndian, uint32(len(dataBytes))); err != nil {
//		return
//	}
//
//	//all pkg data
//	if err = binary.Write(outbuff, binary.LittleEndian, dataBytes); err != nil {
//		return
//	}
//
//	out = outbuff.Bytes()
//	return
//
//}

func (this *RpcDataPack) Pack(msgId uint32, pkg interface{}) (out []byte, err error) {
	outbuff := bytes.NewBuffer([]byte{})
	// 进行编码
	databuff := bytes.NewBuffer([]byte{})
	data := pkg.(*RpcData)
	if data != nil {
		enc := gob.NewEncoder(databuff)
		err = enc.Encode(data)
	}

	if err != nil {
		fmt.Println(fmt.Sprintf("gob marshaling error:  %s", err))
	}
	// 写Len
	if err = binary.Write(outbuff, binary.LittleEndian, uint32(databuff.Len())); err != nil {
		return
	}

	//all pkg data
	if err = binary.Write(outbuff, binary.LittleEndian, databuff.Bytes()); err != nil {
		return
	}

	out = outbuff.Bytes()
	return

}
