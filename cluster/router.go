package cluster

import "github.com/viphxin/xingo/iface"

type BaseRpcRouter struct {
}
func (this *BaseRpcRouter)PreHandle(request iface.IRpcRequest){}
func (this *BaseRpcRouter)Handle(request iface.IRpcRequest) map[string]interface{}{
	return nil
}
func (this *BaseRpcRouter)AfterHandle(request iface.IRpcRequest){}

