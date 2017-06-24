package fnet

/*
	find msg api
*/
import (
	"fmt"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"strconv"
	"time"
	"github.com/viphxin/xingo/iface"
	"runtime/debug"
)

type MsgHandle struct {
	PoolSize  int32
	TaskQueue []chan *Request
	Apis      map[uint32]iface.IRouter
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		PoolSize:  utils.GlobalObject.PoolSize,
		TaskQueue: make([]chan *Request, utils.GlobalObject.PoolSize),
		Apis:      make(map[uint32]iface.IRouter),
	}
}

//一致性路由,保证同一连接的数据转发给相同的goroutine
func (this *MsgHandle) DeliverToMsgQueue(pkg interface{}) {
	request := pkg.(*Request)
	index := request.GetConnection().GetSessionId() % uint32(utils.GlobalObject.PoolSize)
	taskQueue := this.TaskQueue[index]
	logger.Debug(fmt.Sprintf("add to pool : %d", index))
	taskQueue <- request
}

func (this *MsgHandle) DoApi(request *Request){
	defer func() {
		if err := recover(); err != nil {
			logger.Info("-------------DoApi panic recover---------------")
			//this.HandleError(err)
			debug.PrintStack()
		}
	}()
	if f, ok := this.Apis[request.GetMsgId()]; ok {
		//存在
		st := time.Now()
		f.PreHandle(request)
		f.Handle(request)
		f.AfterHandle(request)
		logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", request.GetMsgId(), float64(time.Now().Sub(st).Nanoseconds())/1e6))
	} else {
		logger.Error(fmt.Sprintf("not found api:  %d", request.GetMsgId()))
	}
}

func (this *MsgHandle) DoMsgFromGoRoutine(pkg interface{}) {
	request := pkg.(*Request)
	go func() {
		this.DoApi(request)
	}()
}

func (this *MsgHandle) AddRouter(name string, router iface.IRouter) {
	api := name
	index, err := strconv.Atoi(api)
	if err != nil {
		panic("error api: " + api)
	}
	if _, ok := this.Apis[uint32(index)]; ok {
		//存在
		panic("repeated api " + string(index))
	}
	this.Apis[uint32(index)] = router
	logger.Info("add api " + api)
}

func (this *MsgHandle) AddRpcRouter(name string, router iface.IRpcRouter) {

}

func (this *MsgHandle)HandleError(err interface{}){
        if err != nil{
               logger.Error(err)
        }
}

func (this *MsgHandle) StartWorkerLoop(poolSize int) {
	if utils.GlobalObject.IsThreadSafeMode(){
		//线程安全模式所有的逻辑都在一个goroutine处理, 这样可以实现无锁化服务
		this.TaskQueue[0] = make(chan *Request, utils.GlobalObject.MaxWorkerLen)
		go func(){
			logger.Info("init thread mode workpool.")
			for{
				select {
				case request := <- this.TaskQueue[0]:
					this.DoApi(request)
				case delaytask := <- utils.GlobalObject.GetSafeTimer().GetTriggerChannel():
					delaytask.Call()
				}
			}
		}()
	}else{
		for i := 0; i < poolSize; i += 1 {
			c := make(chan *Request, utils.GlobalObject.MaxWorkerLen)
			this.TaskQueue[i] = c
			go func(index int, taskQueue chan *Request) {
				logger.Info(fmt.Sprintf("init thread pool %d.", index))
				for {
					request := <-taskQueue
					this.DoApi(request)
				}
			}(i, c)
		}
	}
}
