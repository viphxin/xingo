package clusterserver

import (
	"fmt"
	"github.com/viphxin/xingo/cluster"
	"github.com/viphxin/xingo/fnet"
	"github.com/viphxin/xingo/fserver"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"time"
	"syscall"
)

type ClusterServer struct {
	Name           string
	RemoteNodesMgr *cluster.ChildMgr //子节点有
	ChildsMgr      *cluster.ChildMgr //root节点有
	MasterObj      *fnet.TcpClient
	httpServerMux  *http.ServeMux
	NetServer      iface.Iserver
	RootServer     iface.Iserver
	TelnetServer   iface.Iserver
	Cconf          *cluster.ClusterConf
	modules        map[string][]interface{} //所有模块统一管理
	sync.RWMutex
}

func DoCSConnectionLost(fconn iface.Iconnection) {
	logger.Error("node disconnected from " + utils.GlobalObject.Name)
	//子节点掉线
	nodename, err := fconn.GetProperty("child")
	if err == nil {
		GlobalClusterServer.RemoveChild(nodename.(string))
	}
}

func DoCCConnectionLost(fconn iface.Iclient) {
	//父节点掉线
	rname, err := fconn.GetProperty("remote")
	if err == nil {
		GlobalClusterServer.RemoveRemote(rname.(string))
		logger.Error("remote " + rname.(string) + " disconnected from " + utils.GlobalObject.Name)
	}
}

//reconnected to master
func ReConnectMasterCB(fconn iface.Iclient) {
	rpc := cluster.NewChild(utils.GlobalObject.Name, GlobalClusterServer.MasterObj)
	response, err := rpc.CallChildForResult("TakeProxy", utils.GlobalObject.Name)
	if err == nil {
		roots, ok := response.Result["roots"]
		if ok {
			for _, root := range roots.([]string) {
				GlobalClusterServer.ConnectToRemote(root)
			}
		}
	} else {
		panic(fmt.Sprintf("reconnected to master error: %s", err))
	}
}

func NewClusterServer(name, path string) *ClusterServer {
	logger.SetPrefix(fmt.Sprintf("[%s]", strings.ToUpper(name)))
	cconf, err := cluster.NewClusterConf(path)
	if err != nil {
		panic("cluster conf error!!!")
	}

	GlobalClusterServer = &ClusterServer{
		Name:           name,
		Cconf:          cconf,
		RemoteNodesMgr: cluster.NewChildMgr(),
		ChildsMgr:      cluster.NewChildMgr(),
		modules:        make(map[string][]interface{}, 0),
		httpServerMux:  http.NewServeMux(),
	}

	serverconf, ok := GlobalClusterServer.Cconf.Servers[name]
	if !ok {
		panic(fmt.Sprintf("no server %s in clusterconf!!!", name))
	}

	utils.GlobalObject.Name = name
	utils.GlobalObject.OnClusterClosed = DoCSConnectionLost
	utils.GlobalObject.OnClusterCClosed = DoCCConnectionLost
	utils.GlobalObject.RpcCProtoc = cluster.NewRpcClientProtocol()

	if utils.GlobalObject.PoolSize > 0 {
		//init rpc worker pool
		utils.GlobalObject.RpcCProtoc.InitWorker(int32(utils.GlobalObject.PoolSize))
	}
	if serverconf.NetPort > 0 {
		utils.GlobalObject.Protoc = fnet.NewProtocol()

	}
	if serverconf.RootPort > 0 {
		utils.GlobalObject.RpcSProtoc = cluster.NewRpcServerProtocol()
	}

	if serverconf.Log != "" {
		utils.GlobalObject.LogName = serverconf.Log
		utils.ReSettingLog()
	}

	//telnet debug tool
	if serverconf.DebugPort > 0{
		if serverconf.Host != ""{
			GlobalClusterServer.TelnetServer = fserver.NewTcpServer("telnet_server", "tcp4", serverconf.Host, serverconf.DebugPort, 100, cluster.NewTelnetProtocol())
		}else{
			GlobalClusterServer.TelnetServer = fserver.NewTcpServer("telnet_server", "tcp4", "127.0.0.1", serverconf.DebugPort, 100, cluster.NewTelnetProtocol())
		}
	}
	return GlobalClusterServer
}

func (this *ClusterServer) StartClusterServer() {
	serverconf, ok := this.Cconf.Servers[utils.GlobalObject.Name]
	if !ok {
		panic("no server in clusterconf!!!")
	}
	//自动发现注册modules api
	modules, ok := this.modules[serverconf.Module]
	if ok {
		//api
		if serverconf.NetPort > 0 {
			for _, m := range modules[0].([]interface{}){
				if m != nil{
					this.AddRouter(m)
				}
			}
		}
		//http
		if len(serverconf.Http) > 0 || len(serverconf.Https) > 0{
			for _, m := range modules[1].([]interface{}){
				if m != nil{
					this.AddHttpRouter(m)
				}
			}
		}
		//rpc
		for _, m := range modules[2].([]interface{}){
			if m != nil{
				this.AddRpcRouter(m)
			}
		}
	}

	//http server
	if len(serverconf.Http) > 0 {
		//staticfile handel
		if len(serverconf.Http) == 2 {
			this.httpServerMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(serverconf.Http[1].(string)))))
		}
		httpserver := &http.Server{
			Addr:           fmt.Sprintf(":%d", int(serverconf.Http[0].(float64))),
			Handler:        this.httpServerMux,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 20, //1M
		}
		httpserver.SetKeepAlivesEnabled(true)
		go httpserver.ListenAndServe()
		logger.Info(fmt.Sprintf("http://%s:%d start", serverconf.Host, int(serverconf.Http[0].(float64))))
	} else if len(serverconf.Https) > 2 {
		//staticfile handel
		if len(serverconf.Https) == 4 {
			this.httpServerMux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(serverconf.Https[3].(string)))))
		}
		httpserver := &http.Server{
			Addr:           fmt.Sprintf(":%d", int(serverconf.Https[0].(float64))),
			Handler:        this.httpServerMux,
			ReadTimeout:    5 * time.Second,
			WriteTimeout:   5 * time.Second,
			MaxHeaderBytes: 1 << 20, //1M
		}
		httpserver.SetKeepAlivesEnabled(true)
		go httpserver.ListenAndServeTLS(serverconf.Https[1].(string), serverconf.Https[2].(string))
		logger.Info(fmt.Sprintf("http://%s:%d start", serverconf.Host, int(serverconf.Https[0].(float64))))
	}
	//tcp server
	if serverconf.NetPort > 0 {
		utils.GlobalObject.TcpPort = serverconf.NetPort
		if serverconf.Host != ""{
			this.NetServer = fserver.NewTcpServer("xingocluster_net_server", "tcp4", serverconf.Host, serverconf.NetPort,
				utils.GlobalObject.MaxConn, utils.GlobalObject.Protoc)
		}else{
			this.NetServer = fserver.NewTcpServer("xingocluster_net_server", "tcp4", serverconf.Host, serverconf.NetPort,
				utils.GlobalObject.MaxConn, utils.GlobalObject.Protoc)
		}
		this.NetServer.Start()
	}
	if serverconf.RootPort > 0 {
		if serverconf.Host != ""{
			this.RootServer = fserver.NewTcpServer("xingocluster_root_server", "tcp4", serverconf.Host, serverconf.RootPort,
				utils.GlobalObject.IntraMaxConn, utils.GlobalObject.RpcSProtoc)
		}else{
			this.RootServer = fserver.NewTcpServer("xingocluster_root_server", "tcp4", serverconf.Host, serverconf.RootPort,
				utils.GlobalObject.IntraMaxConn, utils.GlobalObject.RpcSProtoc)
		}
		this.RootServer.Start()
	}
	//telnet
	if this.TelnetServer != nil{
		logger.Info(fmt.Sprintf("telnet tool start: %s:%d.", serverconf.Host, serverconf.DebugPort))
		this.TelnetServer.Start()
	}

	//master
	this.ConnectToMaster()

	logger.Info("xingo cluster start success.")
	// close
	this.WaitSignal()
	this.MasterObj.Stop(true)
	if this.RootServer != nil {
		this.RootServer.Stop()
	}

	if this.NetServer != nil {
		this.NetServer.Stop()
	}

	if this.TelnetServer != nil{
		this.TelnetServer.Stop()
	}
	logger.Info("xingo cluster stoped.")
}

func (this *ClusterServer) WaitSignal() {
	signal.Notify(utils.GlobalObject.ProcessSignalChan, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
	sig := <-utils.GlobalObject.ProcessSignalChan
	//尝试主动通知master checkalive
	rpc := cluster.NewChild(utils.GlobalObject.Name, this.MasterObj)
	rpc.CallChildNotForResult("ChildOffLine", utils.GlobalObject.Name)

	logger.Info(fmt.Sprintf("server exit. signal: [%s]", sig))
}

func (this *ClusterServer) ConnectToMaster() {
	master := fnet.NewReConnTcpClient(this.Cconf.Master.Host, this.Cconf.Master.RootPort, utils.GlobalObject.RpcCProtoc, 1024, 60, ReConnectMasterCB)
	this.MasterObj = master
	master.Start()
	//注册到master
	rpc := cluster.NewChild(utils.GlobalObject.Name, this.MasterObj)
	response, err := rpc.CallChildForResult("TakeProxy", utils.GlobalObject.Name)
	if err == nil {
		roots, ok := response.Result["roots"]
		if ok {
			for _, root := range roots.([]string) {
				this.ConnectToRemote(root)
			}
		}
	} else {
		panic(fmt.Sprintf("connected to master error: %s", err))
	}
}

func (this *ClusterServer) ConnectToRemote(rname string) {
	rserverconf, ok := this.Cconf.Servers[rname]
	if ok {
		//处理master掉线，重新通知的情况
		if _, err := this.GetRemote(rname); err != nil {
			rserver := fnet.NewTcpClient(rserverconf.Host, rserverconf.RootPort, utils.GlobalObject.RpcCProtoc)
			this.RemoteNodesMgr.AddChild(rname, rserver)
			rserver.Start()
			rserver.SetProperty("remote", rname)
			//takeproxy
			child, err := this.RemoteNodesMgr.GetChild(rname)
			if err == nil {
				child.CallChildNotForResult("TakeProxy", utils.GlobalObject.Name)
			}
		} else {
			logger.Info("Remote connection already exist!")
		}
	} else {
		//未找到节点
		logger.Error("ConnectToRemote error. " + rname + " node can`t found!!!")
	}
}

func (this *ClusterServer) AddRouter(router interface{}) {
	if utils.GlobalObject.Protoc != nil{
		//add api ---------------start
		utils.GlobalObject.Protoc.AddRpcRouter(router)
		//add api ---------------end
	}
}

func (this *ClusterServer) AddRpcRouter(router interface{}) {
	//add api ---------------start
	utils.GlobalObject.RpcCProtoc.AddRpcRouter(router)
	if utils.GlobalObject.RpcSProtoc != nil {
		utils.GlobalObject.RpcSProtoc.AddRpcRouter(router)
	}
	//add api ---------------end
}

/*
子节点连上来回调
*/
func (this *ClusterServer) AddChild(name string, writer iface.IWriter) {
	this.Lock()
	defer this.Unlock()

	this.ChildsMgr.AddChild(name, writer)
	writer.SetProperty("child", name)
}

/*
子节点断开回调
*/
func (this *ClusterServer) RemoveChild(name string) {
	this.Lock()
	defer this.Unlock()

	this.ChildsMgr.RemoveChild(name)
}

func (this *ClusterServer) RemoveRemote(name string) {
	this.Lock()
	defer this.Unlock()

	this.RemoteNodesMgr.RemoveChild(name)
}

func (this *ClusterServer) GetRemote(name string) (*cluster.Child, error) {
	this.RLock()
	defer this.RUnlock()

	return this.RemoteNodesMgr.GetChild(name)
}

/*
注册模块到分布式服务器
*/
func (this *ClusterServer) AddModule(mname string, apimodule interface{},httpmodule interface{}, rpcmodule interface{}) {
	//this.modules[mname] = []interface{}{module, rpcmodule}
	if _,ok := this.modules[mname]; ok{
		this.modules[mname][0] = append(this.modules[mname][0].([]interface{}), apimodule)
		this.modules[mname][1] = append(this.modules[mname][1].([]interface{}), httpmodule)
		this.modules[mname][2] = append(this.modules[mname][2].([]interface{}), rpcmodule)
	}else{
		this.modules[mname] = []interface{}{[]interface{}{apimodule}, []interface{}{httpmodule}, []interface{}{rpcmodule}}
	}
}

/*
注册http的api到分布式服务器
*/
func (this *ClusterServer) AddHttpRouter(router interface{}) {
	value := reflect.ValueOf(router)
	tp := value.Type()
	for i := 0; i < value.NumMethod(); i += 1 {
		name := tp.Method(i).Name
		uri := fmt.Sprintf("/%s", strings.ToLower(strings.Replace(name, "Handle", "", 1)))
		this.httpServerMux.HandleFunc(uri,
			utils.HttpRequestWrap(uri, value.Method(i).Interface().(func(http.ResponseWriter, *http.Request))))
		logger.Info("add http url: " + uri)
	}
}
