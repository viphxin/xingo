package sys_rpc

import (
	"github.com/viphxin/xingo/cluster"

	"github.com/viphxin/xingo/clusterserver"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"github.com/viphxin/xingo/iface"
)

type RootTakeProxyRouter struct {
	cluster.BaseRpcRouter
}

/*
子节点连上来的通知
*/
func (this *RootTakeProxyRouter) Handle(request iface.IRpcRequest) {
	name := request.GetArgs()[0].(string)
	logger.Info("child node " + name + " connected to " + utils.GlobalObject.Name)
	//加到childs并且绑定链接connetion对象
	clusterserver.GlobalClusterServer.AddChild(name, request.GetWriter())
}
