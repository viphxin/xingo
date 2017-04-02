package cluster

import (
	"errors"
	"fmt"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	_ "os"
	"sync"
	"time"
)

type AsyncResult struct {
	key    string
	result chan *RpcData
}

//func GenUUID() string{
//    f, _ := os.OpenFile("/dev/urandom", os.O_RDONLY, 0)
//    b := make([]byte, 16)
//    f.Read(b)
//    f.Close()
//    return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
//}

var AResultGlobalObj *AsyncResultMgr = NewAsyncResultMgr()

func NewAsyncResult(key string) *AsyncResult {
	return &AsyncResult{
		key:    key,
		result: make(chan *RpcData, 1),
	}
}

func (this *AsyncResult) GetKey() string {
	return this.key
}

func (this *AsyncResult) SetResult(data *RpcData) {
	this.result <- data
}

func (this *AsyncResult) GetResult(timeout time.Duration) (*RpcData, error) {
	select {
	case <-time.After(timeout):
		logger.Error(fmt.Sprintf("GetResult AsyncResult: timeout %s", this.key))
		close(this.result)
		return &RpcData{}, errors.New(fmt.Sprintf("GetResult AsyncResult: timeout %s", this.key))
	case result := <-this.result:
		return result, nil
	}
	return &RpcData{}, errors.New("GetResult AsyncResult error. reason: no")
}

type AsyncResultMgr struct {
	idGen   *utils.UUIDGenerator
	results map[string]*AsyncResult
	sync.RWMutex
}

func NewAsyncResultMgr() *AsyncResultMgr {
	return &AsyncResultMgr{
		results: make(map[string]*AsyncResult, 0),
		idGen:   utils.NewUUIDGenerator("async_result_"),
	}
}

func (this *AsyncResultMgr) Add() *AsyncResult {
	this.Lock()
	defer this.Unlock()

	r := NewAsyncResult(this.idGen.Get())
	this.results[r.GetKey()] = r
	return r
}

func (this *AsyncResultMgr) Remove(key string) {
	this.Lock()
	defer this.Unlock()

	delete(this.results, key)
}

func (this *AsyncResultMgr) GetAsyncResult(key string) (*AsyncResult, error) {
	this.RLock()
	defer this.RUnlock()

	r, ok := this.results[key]
	if ok {
		return r, nil
	} else {
		return nil, errors.New("not found AsyncResult")
	}
}

func (this *AsyncResultMgr) FillAsyncResult(key string, data *RpcData) error {
	r, err := this.GetAsyncResult(key)
	if err == nil {
		this.Remove(key)
		r.SetResult(data)
		return nil
	} else {
		return err
	}
}
