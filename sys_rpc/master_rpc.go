package sys_rpc

import (
	"github.com/viphxin/xingo/cluster"
	"github.com/viphxin/xingo/clusterserver"
	"github.com/viphxin/xingo/logger"
)

type MasterRpc struct {
}

func (this *MasterRpc) TakeProxy(request *cluster.RpcRequest) (response map[string]interface{}) {
	response = make(map[string]interface{}, 0)
	name := request.Rpcdata.Args[0].(string)
	logger.Info("node " + name + " connected to master.")
	//加到childs并且绑定链接connetion对象
	clusterserver.GlobalMaster.AddNode(name, request.Fconn)

	//返回需要链接的父节点
	remotes, err := clusterserver.GlobalMaster.Cconf.GetRemotesByName(name)
	if err == nil {
		roots := make([]string, 0)
		for _, r := range remotes {
			if _, ok := clusterserver.GlobalMaster.OnlineNodes[r]; ok {
				//父节点在线
				roots = append(roots, r)
			}
		}
		response["roots"] = roots
	}
	//通知当前节点的子节点链接当前节点
	for _, child := range clusterserver.GlobalMaster.Childs.GetChilds() {
		//遍历所有子节点,观察child节点的父节点是否包含当前节点
		remotes, err := clusterserver.GlobalMaster.Cconf.GetRemotesByName(child.GetName())
		if err == nil {
			for _, rname := range remotes {
				if rname == name {
					//包含，需要通知child节点连接当前节点
					//rpc notice
					child.CallChildNotForResult("RootTakeProxy", name)
					break
				}
			}
		}
	}
	return
}

//主动通知master 节点掉线
func (this *MasterRpc) ChildOffLine(request *cluster.RpcRequest) {
	name := request.Rpcdata.Args[0].(string)
	logger.Info("node " + name + " disconnected offline.")
	clusterserver.GlobalMaster.CheckChildsAlive(true)
}
