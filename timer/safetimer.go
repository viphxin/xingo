package timer

import (
	"github.com/viphxin/xingo/logger"
	"sync"
	"time"
	"math"
)

/*
协程安全的timer
*/
const (
	//默认安全时间调度器的容量
	TIMERLEN = 2048
	//默认最大误差值100毫秒
	ERRORMAX = 100
	//默认最大触发队列缓冲大小
	TRIGGERMAX = 2048
	//默认hashwheel分级
	LEVEL     =  12
)

func UnixTS() int64 {
	return time.Now().UnixNano() / 1e6
}

type ParamNull struct {}

type SafeTimer struct {
	//延迟调用的函数
	delayCall *DelayCall
	//调用的时间：单位毫秒
	unixts int64
}

func NewSafeTimer(delay int64, delayCall *DelayCall) *SafeTimer {
	unixts := UnixTS()
	if delay > 0 {
		unixts += delay
	}
	return &SafeTimer{
		delayCall: delayCall,
		unixts:    unixts,
	}
}

type SafeTimerScheduel struct {
	hashwheel   *HashWheel
	idGen       uint32
	triggerChan chan *DelayCall
	sync.RWMutex
}

func NewSafeTimerScheduel() *SafeTimerScheduel {
	scheduel := &SafeTimerScheduel{
		hashwheel:       NewHashWheel("wheel_hours", LEVEL, 3600*1e3, TIMERLEN),
		idGen:       0,
		triggerChan: make(chan *DelayCall, TRIGGERMAX),
	}

	//minute wheel
	minuteWheel := NewHashWheel("wheel_minutes", LEVEL, 60*1e3, TIMERLEN)
	//second wheel
	secondWheel := NewHashWheel("wheel_seconds", LEVEL, 1*1e3, TIMERLEN)
	minuteWheel.AddNext(secondWheel)
	scheduel.hashwheel.AddNext(minuteWheel)

	go scheduel.StartScheduelLoop()
	return scheduel
}

func (this *SafeTimerScheduel) GetTriggerChannel() chan *DelayCall {
	return this.triggerChan
}

func (this *SafeTimerScheduel) CreateTimer(delay int64, f func(v ...interface{}), args []interface{}) (uint32, error) {
	this.Lock()
	defer this.Unlock()

	this.idGen += 1
	err := this.hashwheel.Add2WheelChain(this.idGen,
		NewSafeTimer(delay, &DelayCall{
			f:    f,
			args: args,
		}))
	if err != nil{
		return 0, err
	}else{
		return this.idGen, nil
	}
}

func (this *SafeTimerScheduel) CancelTimer(timerId uint32) {
	this.hashwheel.RemoveFromWheelChain(timerId)
}

func (this *SafeTimerScheduel) StartScheduelLoop() {
	logger.Info("xingo safe timer scheduelloop runing.")
	for {
		triggerList := this.hashwheel.GetTriggerWithIn(ERRORMAX)
		//trigger
		for _, v := range triggerList {
			//logger.Debug("want call: ", v.unixts, ".real call: ", UnixTS(), ".ErrorMS: ", UnixTS()-v.unixts)
			if math.Abs(float64(UnixTS()-v.unixts)) > float64(ERRORMAX){
				logger.Error("want call: ", v.unixts, ".real call: ", UnixTS(), ".ErrorMS: ", UnixTS()-v.unixts)
			}
			this.triggerChan <- v.delayCall
		}

		//wait for next loop
		time.Sleep(ERRORMAX/2*time.Millisecond)
	}
}