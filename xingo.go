package xingo

import (
	_ "github.com/viphxin/xingo/fnet"
	_ "github.com/viphxin/xingo/timer"
	"github.com/viphxin/xingo/telnetcmd"
	"github.com/viphxin/xingo/clusterserver"
	"github.com/viphxin/xingo/sys_rpc"
	"github.com/viphxin/xingo/utils"
	"github.com/viphxin/xingo/fserver"
	"github.com/viphxin/xingo/cluster"
	"github.com/viphxin/xingo/logger"
	"fmt"
	"github.com/viphxin/xingo/iface"
)

func NewXingoTcpServer() iface.Iserver{
	//do something
	//debugport 是否开放
	if utils.GlobalObject.DebugPort > 0{
		if utils.GlobalObject.Host != ""{
			fserver.NewTcpServer("telnet_server", "tcp4", utils.GlobalObject.Host,
				utils.GlobalObject.DebugPort, 100, cluster.NewTelnetProtocol()).Start()
		}else{
			fserver.NewTcpServer("telnet_server", "tcp4", "127.0.0.1",
				utils.GlobalObject.DebugPort, 100, cluster.NewTelnetProtocol()).Start()
		}
		logger.Debug(fmt.Sprintf("telnet tool start: %s:%d.", utils.GlobalObject.Host, utils.GlobalObject.DebugPort))

	}

	//add command
	if utils.GlobalObject.CmdInterpreter != nil{
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewPprofCpuCommand())
	}

	s := fserver.NewServer()
	return s
}

func NewXingoMaster(cfg string) *clusterserver.Master{
	s := clusterserver.NewMaster(cfg)
	//add rpc
	s.AddRpcRouter("MasterTakeProxy", &sys_rpc.MasterTakeProxyRouter{})
	s.AddRpcRouter("ChildOffLine", &sys_rpc.ChildOffLineRouter{})
	//add command
	if utils.GlobalObject.CmdInterpreter != nil{
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewPprofCpuCommand())
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewCloseServerCommand())
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewReloadCfgCommand())
	}
	return s
}

func NewXingoCluterServer(nodename, cfg string) *clusterserver.ClusterServer{
	s := clusterserver.NewClusterServer(nodename,cfg)
	//add rpc
	s.AddRpcRouter("NtfTakeProxy", &sys_rpc.NtfTakeProxyRouter{})
	s.AddRpcRouter("CloseServer", &sys_rpc.CloseServerRouter{})
	s.AddRpcRouter("ReloadConfig", &sys_rpc.ReloadConfigRouter{})
	s.AddRpcRouter("RootTakeProxy", &sys_rpc.RootTakeProxyRouter{})
	s.AddRpcRouter("CheckAlive", &sys_rpc.CheckAliveRouter{})
	s.AddRpcRouter("NodeDownNtf", &sys_rpc.NodeDownNtfRouter{})
	//add cmd
	if utils.GlobalObject.CmdInterpreter != nil{
		utils.GlobalObject.CmdInterpreter.AddCommand(telnetcmd.NewPprofCpuCommand())
	}
	return s
}
