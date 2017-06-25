package cluster

/*
	regest rpc
*/
import (
	"fmt"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"math/rand"
	"time"
	"github.com/viphxin/xingo/iface"
	"runtime/debug"
)

type RpcMsgHandle struct {
	PoolSize  int32
	TaskQueue []chan *RpcRequest
	Apis      map[string]iface.IRpcRouter
}

func NewRpcMsgHandle() *RpcMsgHandle {
	return &RpcMsgHandle{
		PoolSize:  utils.GlobalObject.PoolSize,
		TaskQueue: make([]chan *RpcRequest, utils.GlobalObject.PoolSize),
		Apis:      make(map[string]iface.IRpcRouter),
	}
}

/*
处理rpc消息
*/
func (this *RpcMsgHandle) DoMsg(request *RpcRequest) {
	defer func() {
		if err := recover(); err != nil {
			logger.Info("-------------DoRpcMsg panic recover---------------")
			logger.Error(err)
			debug.PrintStack()
		}
	}()
	if request.Rpcdata.MsgType == RESPONSE && request.Rpcdata.Key != "" {
		//放回异步结果
		AResultGlobalObj.FillAsyncResult(request.Rpcdata.Key, request.Rpcdata)
		return
	} else {
		//rpc 请求
		if f, ok := this.Apis[request.Rpcdata.Target]; ok {
			//存在
			st := time.Now()
			f.PreHandle(request)
			if request.Rpcdata.MsgType == REQUEST_FORRESULT {
				f.Handle(request)
				packdata, err := utils.GlobalObject.RpcCProtoc.GetDataPack().Pack(0, &RpcData{
					MsgType: RESPONSE,
					Result:  request.GetResult(),
					Key:     request.Rpcdata.Key,
				})
				if err == nil {
					request.Fconn.Send(packdata)
				} else {
					logger.Error(err)
				}
			} else if request.Rpcdata.MsgType == REQUEST_NORESULT {
				f.Handle(request)
			}
			f.AfterHandle(request)
			logger.Debug(fmt.Sprintf("rpc %s cost total time: %f ms", request.Rpcdata.Target, float64(time.Now().Sub(st).Nanoseconds())/1e6))
		} else {
			logger.Error(fmt.Sprintf("not found rpc:  %s", request.Rpcdata.Target))
		}
	}
}

func (this *RpcMsgHandle) DeliverToMsgQueue(pkg interface{}) {
	request := pkg.(*RpcRequest)
	//add to worker pool
	index := rand.Int31n(utils.GlobalObject.PoolSize)
	taskQueue := this.TaskQueue[index]
	logger.Debug(fmt.Sprintf("add to rpc pool : %d", index))
	taskQueue <- request
}

func (this *RpcMsgHandle) DoMsgFromGoRoutine(pkg interface{}) {
	request := pkg.(*RpcRequest)
	go this.DoMsg(request)
}

func (this *RpcMsgHandle) AddRpcRouter(name string, router iface.IRpcRouter) {
	if _, ok := this.Apis[name]; ok {
		//存在
		panic("repeated rpc " + name)
	}
	this.Apis[name] = router
	logger.Info("add rpc " + name)
}
func (this *RpcMsgHandle) AddRouter(name string, router iface.IRouter) {
	//don`t need implement
}

func (this *RpcMsgHandle) StartWorkerLoop(poolSize int) {
	if utils.GlobalObject.IsThreadSafeMode(){
		this.TaskQueue[0] =  make(chan *RpcRequest, utils.GlobalObject.MaxWorkerLen)
		go func(){
			for{
				select {
				case rpcRequest := <- this.TaskQueue[0]:
					this.DoMsg(rpcRequest)
				case delayCall := <- utils.GlobalObject.GetSafeTimer().GetTriggerChannel():
					delayCall.Call()
				}
			}
		}()
	}else{
		for i := 0; i < poolSize; i += 1 {
			c := make(chan *RpcRequest, utils.GlobalObject.MaxWorkerLen)
			this.TaskQueue[i] = c
			go func(index int, taskQueue chan *RpcRequest) {
				logger.Info(fmt.Sprintf("init rpc thread pool %d.", index))
				for {
					request := <-taskQueue
					this.DoMsg(request)
				}

			}(i, c)
		}
	}
}
