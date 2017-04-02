package fnet

/*
	find msg api
*/
import (
	"fmt"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type MsgHandle struct {
	PoolSize  int32
	TaskQueue []chan *PkgAll
	Apis      map[uint32]reflect.Value
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		PoolSize:  utils.GlobalObject.PoolSize,
		TaskQueue: make([]chan *PkgAll, utils.GlobalObject.PoolSize),
		Apis:      make(map[uint32]reflect.Value),
	}
}

func (this *MsgHandle) DeliverToMsgQueue(pkg interface{}) {
	data := pkg.(*PkgAll)
	index := rand.Int31n(utils.GlobalObject.PoolSize)
	taskQueue := this.TaskQueue[index]
	logger.Debug(fmt.Sprintf("add to pool : %d", index))
	taskQueue <- data
}

func (this *MsgHandle) DoMsgFromGoRoutine(pkg interface{}) {
	data := pkg.(*PkgAll)
	go func() {
		if f, ok := this.Apis[data.Pdata.MsgId]; ok {
			//存在
			st := time.Now()
			f.Call([]reflect.Value{reflect.ValueOf(data)})
			logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", data.Pdata.MsgId, time.Now().Sub(st).Seconds()*1000))
		} else {
			logger.Error(fmt.Sprintf("not found api:  %d", data.Pdata.MsgId))
		}
	}()
}

func (this *MsgHandle) AddRouter(router interface{}) {
	value := reflect.ValueOf(router)
	tp := value.Type()
	for i := 0; i < value.NumMethod(); i += 1 {
		name := tp.Method(i).Name
		k := strings.Split(name, "_")
		index, err := strconv.Atoi(k[1])
		if err != nil {
			panic("error api: " + name)
		}
		if _, ok := this.Apis[uint32(index)]; ok {
			//存在
			panic("repeated api " + string(index))
		}
		this.Apis[uint32(index)] = value.Method(i)
		logger.Info("add api " + name)
	}

	//exec test
	// for i := 0; i < 100; i += 1 {
	// 	Apis[1].Call([]reflect.Value{reflect.ValueOf("huangxin"), reflect.ValueOf(router)})
	// 	Apis[2].Call([]reflect.Value{})
	// }
	// fmt.Println(this.Apis)
	// this.Apis[2].Call([]reflect.Value{reflect.ValueOf(&PkgAll{})})
}

func (this *MsgHandle) StartWorkerLoop(poolSize int) {
	for i := 0; i < poolSize; i += 1 {
		c := make(chan *PkgAll, utils.GlobalObject.MaxWorkerLen)
		this.TaskQueue[i] = c
		go func(index int, taskQueue chan *PkgAll) {
			logger.Info(fmt.Sprintf("init thread pool %d.", index))
			for {
				data := <-taskQueue
				//can goroutine?
				if f, ok := this.Apis[data.Pdata.MsgId]; ok {
					//存在
					st := time.Now()
					//f.Call([]reflect.Value{reflect.ValueOf(data)})
					utils.XingoTry(f, []reflect.Value{reflect.ValueOf(data)}, nil)
					logger.Debug(fmt.Sprintf("Api_%d cost total time: %f ms", data.Pdata.MsgId, time.Now().Sub(st).Seconds()*1000))
				} else {
					logger.Error(fmt.Sprintf("not found api:  %d", data.Pdata.MsgId))
				}
			}
		}(i, c)
	}
}
