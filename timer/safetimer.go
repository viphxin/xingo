package timer

import (
	"github.com/viphxin/xingo/logger"
	"sync"
	"time"
)

/*
协程安全的timer
*/
const (
	//默认安全时间调度器的容量
	TIMERLEN = 2048
	//默认最大误差值10毫秒
	ERRORMAX = 10
	//默认最大触发队列缓冲大小
	TRIGGERMAX = 1024
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
	//最早需要触发的timer
	min         int64
	tasks       map[uint32]*SafeTimer
	idGen       uint32
	triggerChan chan *DelayCall
	sync.RWMutex
}

func NewSafeTimerScheduel() *SafeTimerScheduel {
	scheduel := &SafeTimerScheduel{
		tasks:       make(map[uint32]*SafeTimer, TIMERLEN),
		idGen:       0,
		triggerChan: make(chan *DelayCall, TRIGGERMAX),
	}
	go scheduel.StartScheduelLoop()
	return scheduel
}

func (this *SafeTimerScheduel) GetTriggerChannel() chan *DelayCall {
	return this.triggerChan
}

func (this *SafeTimerScheduel) add(delayCall *SafeTimer) uint32 {
	this.idGen += 1
	this.tasks[this.idGen] = delayCall
	if this.min > delayCall.unixts {
		this.min = delayCall.unixts
	}
	return this.idGen
}

func (this *SafeTimerScheduel) remove(timerId uint32) {
	if _, ok := this.tasks[timerId]; ok {
		delete(this.tasks, timerId)
	}
}

func (this *SafeTimerScheduel) CreateTimer(delay int64, f func(v ...interface{}), args []interface{}) uint32 {
	this.Lock()
	defer this.Unlock()

	return this.add(
		NewSafeTimer(delay, &DelayCall{
			f:    f,
			args: args,
		}))
}

func (this *SafeTimerScheduel) CancelTimer(timerId uint32) {
	this.Lock()
	defer this.Unlock()

	this.remove(timerId)
}

func (this *SafeTimerScheduel) StartScheduelLoop() {
	logger.Info("xingo safe timer scheduelloop runing.")
	for {
		unixts := UnixTS()
		if this.min-unixts > ERRORMAX {
			//如果最早需要触发的定时器大于ERRORMAX ms
			time.Sleep(ERRORMAX * time.Millisecond)
		}
		triggerList := make(map[int8][]*SafeTimer, ERRORMAX)
		timerIdMap := make([]uint32, 0)
		this.Lock()
		for k, v := range this.tasks {
			if v.unixts-unixts > 10 {
				continue
			} else {
				x := v.unixts - unixts
				if x <= 0 {
					x = 1
				}
				x -= 1

				if xList, ok := triggerList[int8(x)]; ok {
					xList = append(xList, v)
				} else {
					xList = []*SafeTimer{v}
					triggerList[int8(x)] = xList
				}
				timerIdMap = append(timerIdMap, k)
			}
		}
		//delete
		for _, timerId := range timerIdMap {
			this.remove(timerId)
		}
		this.Unlock()
		//trigger
		if len(triggerList) > 0 {
			logger.Debug(triggerList)
		}
		for _, v := range triggerList {
			for _, t := range v {
				logger.Debug("want call: ", t.unixts, ".real call: ", UnixTS(), ".ErrorMS: ", UnixTS()-t.unixts)
				this.triggerChan <- t.delayCall
			}
		}

		//wait for next loop
		time.Sleep(ERRORMAX * time.Millisecond)
	}
}