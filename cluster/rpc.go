package cluster

import (
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"time"
)

type RpcSignal int32

const (
	REQUEST_NORESULT RpcSignal = iota
	REQUEST_FORRESULT
	RESPONSE
)

type XingoRpc struct {
	conn           iface.IWriter
	asyncResultMgr *AsyncResultMgr
}

func NewXingoRpc(conn iface.IWriter) *XingoRpc {
	return &XingoRpc{
		conn:           conn,
		asyncResultMgr: AResultGlobalObj,
	}
}

func (this *XingoRpc) CallRpcNotForResult(target string, args ...interface{}) error {
	rpcdata := &RpcData{
		MsgType: REQUEST_NORESULT,
		Target:  target,
		Args:    args,
	}
	rpcpackege, err := utils.GlobalObject.RpcCProtoc.GetDataPack().Pack(0, rpcdata)

	if err == nil {
		this.conn.Send(rpcpackege)
		return nil
	} else {
		logger.Error(err)
		return err
	}
}

func (this *XingoRpc) CallRpcForResult(target string, args ...interface{}) (*RpcData, error) {
	asyncR := this.asyncResultMgr.Add()
	rpcdata := &RpcData{
		MsgType: REQUEST_FORRESULT,
		Key:     asyncR.GetKey(),
		Target:  target,
		Args:    args,
	}
	rpcpackege, err := utils.GlobalObject.RpcCProtoc.GetDataPack().Pack(0, rpcdata)
	if err == nil {
		this.conn.Send(rpcpackege)
		resp, err := asyncR.GetResult(5 * time.Second)
		if err == nil {
			return resp, nil
		} else {
			//超时了 或者其他原因结果没等到
			this.asyncResultMgr.Remove(asyncR.GetKey())
			return nil, err
		}
	} else {
		logger.Error(err)
		return nil, err
	}
}
