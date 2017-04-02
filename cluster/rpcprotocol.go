package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"io"
)

type RpcServerProtocol struct {
	rpcMsgHandle *RpcMsgHandle
	rpcDatapack  *RpcDataPack
}

func NewRpcServerProtocol() *RpcServerProtocol {
	return &RpcServerProtocol{
		rpcMsgHandle: NewRpcMsgHandle(),
		rpcDatapack:  NewRpcDataPack(),
	}
}

func (this *RpcServerProtocol) GetMsgHandle() iface.Imsghandle {
	return this.rpcMsgHandle
}

func (this *RpcServerProtocol) GetDataPack() iface.Idatapack {
	return this.rpcDatapack
}

func (this *RpcServerProtocol) AddRpcRouter(router interface{}) {
	this.rpcMsgHandle.AddRouter(router)
}

func (this *RpcServerProtocol) InitWorker(poolsize int32) {
	this.rpcMsgHandle.StartWorkerLoop(int(poolsize))
}

func (this *RpcServerProtocol) OnConnectionMade(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("node ID: %d connected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClusterConnectioned(fconn)
}

func (this *RpcServerProtocol) OnConnectionLost(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("node ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClusterClosed(fconn)
}

func (this *RpcServerProtocol) StartReadThread(fconn iface.Iconnection) {
	logger.Debug("start receive rpc data from socket...")
	for {
		//read per head data
		headdata := make([]byte, this.rpcDatapack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := this.rpcDatapack.Unpack(headdata)
		if err != nil {
			fconn.Stop()
			return
		}
		//data
		pkg := pkgHead.(*RpcPackege)
		if pkg.Len > 0 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				fconn.Stop()
				return
			} else {
				rpcRequest := &RpcRequest{
					Fconn:   fconn,
					Rpcdata: &RpcData{},
				}

				err = json.Unmarshal(pkg.Data, rpcRequest.Rpcdata)

				if err != nil {
					logger.Error(err)
					fconn.Stop()
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkg.Len, int(rpcRequest.Rpcdata.MsgType)))
				if utils.GlobalObject.IsUsePool && rpcRequest.Rpcdata.MsgType != RESPONSE {
					this.rpcMsgHandle.DeliverToMsgQueue(rpcRequest)
				} else {
					this.rpcMsgHandle.DoMsgFromGoRoutine(rpcRequest)
				}
			}
		}
	}
}

type RpcClientProtocol struct {
	rpcMsgHandle *RpcMsgHandle
	rpcDatapack  *RpcDataPack
}

func NewRpcClientProtocol() *RpcClientProtocol {
	return &RpcClientProtocol{
		rpcMsgHandle: NewRpcMsgHandle(),
		rpcDatapack:  NewRpcDataPack(),
	}
}

func (this *RpcClientProtocol) GetMsgHandle() iface.Imsghandle {
	return this.rpcMsgHandle
}

func (this *RpcClientProtocol) GetDataPack() iface.Idatapack {
	return this.rpcDatapack
}
func (this *RpcClientProtocol) AddRpcRouter(router interface{}) {
	this.rpcMsgHandle.AddRouter(router)
}

func (this *RpcClientProtocol) InitWorker(poolsize int32) {
	this.rpcMsgHandle.StartWorkerLoop(int(poolsize))
}

func (this *RpcClientProtocol) OnConnectionMade(fconn iface.Iclient) {
	utils.GlobalObject.OnClusterCConnectioned(fconn)
}

func (this *RpcClientProtocol) OnConnectionLost(fconn iface.Iclient) {
	utils.GlobalObject.OnClusterCClosed(fconn)
}

func (this *RpcClientProtocol) StartReadThread(fconn iface.Iclient) {
	logger.Debug("start receive rpc data from socket...")
	for {
		//read per head data
		headdata := make([]byte, this.rpcDatapack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop(false)
			return
		}
		pkgHead, err := this.rpcDatapack.Unpack(headdata)
		if err != nil {
			fconn.Stop(false)
			return
		}
		//data
		pkg := pkgHead.(*RpcPackege)
		if pkg.Len > 0 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				fconn.Stop(false)
				return
			} else {
				rpcRequest := &RpcRequest{
					Fconn:   fconn,
					Rpcdata: &RpcData{},
				}
				err = json.Unmarshal(pkg.Data, rpcRequest.Rpcdata)
				if err != nil {
					logger.Error("json.Unmarshal error!!!")
					fconn.Stop(false)
					return
				}

				logger.Debug(fmt.Sprintf("rpc call. data len: %d. MsgType: %d", pkg.Len, rpcRequest.Rpcdata.MsgType))
				if utils.GlobalObject.IsUsePool && rpcRequest.Rpcdata.MsgType != RESPONSE {
					this.rpcMsgHandle.DeliverToMsgQueue(rpcRequest)
				} else {
					this.rpcMsgHandle.DoMsgFromGoRoutine(rpcRequest)
				}
			}
		}
	}
}
