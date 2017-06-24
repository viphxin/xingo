package iface

import (
	"time"
)

type Iserver interface {
	Start()
	Stop()
	Serve()
	GetConnectionMgr() Iconnectionmgr
	GetConnectionQueue() chan interface{}
	AddRouter(name string, router IRouter)
	CallLater(durations time.Duration, f func(v ...interface{}), args ...interface{})
	CallWhen(ts string, f func(v ...interface{}), args ...interface{})
	CallLoop(durations time.Duration, f func(v ...interface{}), args ...interface{})
}
