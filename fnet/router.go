package fnet

import (
	"github.com/viphxin/xingo/iface"
	"net/http"
)

type BaseRouter struct {
}

func (this *BaseRouter)PreHandle(request iface.IRequest){}
func (this *BaseRouter)Handle(request iface.IRequest){}
func (this *BaseRouter)AfterHandle(request iface.IRequest){}

type BaseHttpRouter struct {
}

func (this *BaseHttpRouter)PreHandle(response http.ResponseWriter, request *http.Request){}
func (this *BaseHttpRouter)Handle(response http.ResponseWriter, request *http.Request){}
func (this *BaseHttpRouter)AfterHandle(response http.ResponseWriter, request *http.Request){}
