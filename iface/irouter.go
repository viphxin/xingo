package iface

import "net/http"

type IRequest interface {
	GetConnection() Iconnection
	GetData() []byte
	GetMsgId() uint32
}

type IRouter interface {
	PreHandle(IRequest)
	Handle(IRequest)
	AfterHandle(IRequest)
}

type IRpcRequest interface {
	GetWriter() IWriter
	GetMsgType() int32
	GetKey() string
	GetTarget() string
	GetArgs() []interface{}
	GetResult() map[string]interface{}
	PushReturn(string, interface{})
}

type IRpcRouter interface {
	PreHandle(IRpcRequest)
	Handle(IRpcRequest)
	AfterHandle(IRpcRequest)
}

type IHttpRouter interface {
	PreHandle(http.ResponseWriter, *http.Request)
	Handle(http.ResponseWriter, *http.Request)
	AfterHandle(http.ResponseWriter, *http.Request)
}
