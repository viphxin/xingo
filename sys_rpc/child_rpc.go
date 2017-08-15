package sys_rpc

import (
	"fmt"
	"github.com/viphxin/xingo/cluster"
	"github.com/viphxin/xingo/clusterserver"
	"github.com/viphxin/xingo/logger"
	"time"
	"github.com/viphxin/xingo/utils"
	"os"
	"github.com/viphxin/xingo/iface"
)

type NtfTakeProxyRouter struct {
	cluster.BaseRpcRouter
}

/*
master 通知父节点上线, 收到通知的子节点需要链接对应父节点
*/
func (this *NtfTakeProxyRouter) Handle(request iface.IRpcRequest) {
	rname := request.GetArgs()[0].(string)
	logger.Info(fmt.Sprintf("root node %s online. connecting...", rname))
	clusterserver.GlobalClusterServer.ConnectToRemote(rname)
}

type CloseServerRouter struct {
	cluster.BaseRpcRouter
}
/*
关闭节点信号
*/
func (this *CloseServerRouter) Handle(request iface.IRpcRequest){
	delay := request.GetArgs()[0].(int)
	logger.Warn("server close kickdown.", delay, "second...")
	time.Sleep(time.Duration(delay)*time.Second)
	utils.GlobalObject.ProcessSignalChan <- os.Kill
}


type ReloadConfigRouter struct {
	cluster.BaseRpcRouter
}
/*
重新加载配置文件
*/
func (this *ReloadConfigRouter) Handle(request iface.IRpcRequest){
	delay := request.GetArgs()[0].(int)
	logger.Warn("server ReloadConfig kickdown.", delay, "second...")
	time.Sleep(time.Duration(delay)*time.Second)
	clusterserver.GlobalClusterServer.Cconf.Reload()
	logger.Info("reload config.")
}
