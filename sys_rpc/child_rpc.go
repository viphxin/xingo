package sys_rpc

import (
	"fmt"
	"github.com/viphxin/xingo/cluster"
	"github.com/viphxin/xingo/clusterserver"
	"github.com/viphxin/xingo/logger"
	"time"
	"github.com/viphxin/xingo/utils"
	"os"
)

type ChildRpc struct {
}

/*
master 通知父节点上线, 收到通知的子节点需要链接对应父节点
*/
func (this *ChildRpc) RootTakeProxy(request *cluster.RpcRequest) {
	rname := request.Rpcdata.Args[0].(string)
	logger.Info(fmt.Sprintf("root node %s online. connecting...", rname))
	clusterserver.GlobalClusterServer.ConnectToRemote(rname)
}

/*
关闭节点信号
*/
func (this *ChildRpc) CloseServer(request *cluster.RpcRequest){
	delay := request.Rpcdata.Args[0].(int)
	logger.Warn("server close kickdown.", delay, "second...")
	time.Sleep(time.Duration(delay)*time.Second)
	utils.GlobalObject.ProcessSignalChan <- os.Kill
}

/*
重新加载配置文件
*/
func (this *ChildRpc) ReloadConfig(request *cluster.RpcRequest){
	delay := request.Rpcdata.Args[0].(int)
	logger.Warn("server ReloadConfig kickdown.", delay, "second...")
	time.Sleep(time.Duration(delay)*time.Second)
	clusterserver.GlobalClusterServer.Cconf.Reload()
	utils.GlobalObject.Reload()
	logger.Info("reload config.")
}


/*
检查节点是否下线
*/
func (this *ChildRpc) CheckAlive(request *cluster.RpcRequest)(response map[string]interface{}){
	logger.Debug("CheckAlive!")
	response = make(map[string]interface{})
	response["name"] = clusterserver.GlobalClusterServer.Name
	return
}

/*
通知节点掉线（父节点或子节点）
*/
func (this *ChildRpc)NodeDownNtf(request *cluster.RpcRequest) {
	isChild := request.Rpcdata.Args[0].(bool)
	nodeName := request.Rpcdata.Args[1].(string)
	logger.Debug(fmt.Sprintf("node %s down ntf.", nodeName))
	if isChild {
		clusterserver.GlobalClusterServer.RemoveChild(nodeName)
	}else{
		clusterserver.GlobalClusterServer.RemoveRemote(nodeName)
	}
}