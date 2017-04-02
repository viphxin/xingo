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
)

type ClusterServer struct {
	Name           string
	RemoteNodesMgr *cluster.ChildMgr //子节点有
	ChildsMgr      *cluster.ChildMgr //root节点有
	MasterObj      *fnet.TcpClient
	httpServerMux  *http.ServeMux
	RootServer     iface.Iserver
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
			for _, root := range roots.([]interface{}) {
				GlobalClusterServer.ConnectToRemote(root.(string))
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
	utils.GlobalObject.Name = name
	utils.GlobalObject.OnClusterClosed = DoCSConnectionLost
	utils.GlobalObject.OnClusterCClosed = DoCCConnectionLost
	utils.GlobalObject.RpcCProtoc = cluster.NewRpcClientProtocol()

	if GlobalClusterServer.Cconf.Servers[name].NetPort != 0 {
		utils.GlobalObject.Protoc = fnet.NewProtocol()
		if utils.GlobalObject.IsUsePool {
			//init rpc worker pool
			utils.GlobalObject.RpcCProtoc.InitWorker(int32(utils.GlobalObject.PoolSize))
		}
	} else {
		utils.GlobalObject.RpcSProtoc = cluster.NewRpcServerProtocol()
		utils.GlobalObject.Protoc = utils.GlobalObject.RpcSProtoc
	}

	if GlobalClusterServer.Cconf.Servers[name].Log != "" {
		utils.GlobalObject.LogName = GlobalClusterServer.Cconf.Servers[name].Log
		utils.ReSettingLog()
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
		if modules[0] != nil {
			if len(serverconf.Http) > 0 || len(serverconf.Https) > 0 {
				this.AddHttpRouter(modules[0])
			} else {
				this.AddRouter(modules[0])
			}
		}

		if modules[1] != nil {
			this.AddRpcRouter(modules[1])
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
		this.RootServer = fserver.NewServer()
		this.RootServer.Start()
	} else if serverconf.RootPort > 0 {
		utils.GlobalObject.TcpPort = serverconf.RootPort
		this.RootServer = fserver.NewServer()
		this.RootServer.Start()
	}
	//master
	this.ConnectToMaster()

	logger.Info("xingo cluster start success.")
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	<-c
	this.MasterObj.Stop(true)
	if this.RootServer != nil {
		this.RootServer.Stop()
	}
	logger.Info("xingo cluster stoped.")
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
			for _, root := range roots.([]interface{}) {
				this.ConnectToRemote(root.(string))
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
	//add api ---------------start
	utils.GlobalObject.Protoc.AddRpcRouter(router)
	//add api ---------------end
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
func (this *ClusterServer) AddModule(mname string, module interface{}, rpcmodule interface{}) {
	this.modules[mname] = []interface{}{module, rpcmodule}
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
