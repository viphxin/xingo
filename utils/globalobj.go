package utils

import (
	"encoding/json"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"io/ioutil"
)

type GlobalObj struct {
	TcpServer              iface.Iserver
	OnConnectioned         func(fconn iface.Iconnection)
	OnClosed               func(fconn iface.Iconnection)
	OnClusterConnectioned  func(fconn iface.Iconnection) //集群rpc root节点回调
	OnClusterClosed        func(fconn iface.Iconnection)
	OnClusterCConnectioned func(fconn iface.Iclient) //集群rpc 子节点回调
	OnClusterCClosed       func(fconn iface.Iclient)
	OnServerStop           func() //服务器停服回调
	Protoc                 iface.IServerProtocol
	RpcSProtoc             iface.IServerProtocol
	RpcCProtoc             iface.IClientProtocol
	TcpPort                int
	MaxConn                int
	//log
	LogPath        string
	LogName        string
	MaxLogNum      int32
	MaxFileSize    int64
	LogFileUnit    logger.UNIT
	LogLevel       logger.LEVEL
	SetToConsole   bool
	LogFileType    int32
	PoolSize       int32
	IsUsePool      bool
	MaxWorkerLen   int32
	MaxSendChanLen int32
	FrameSpeed     uint8
	Name           string
}

var GlobalObject *GlobalObj

func init() {
	GlobalObject = &GlobalObj{
		TcpPort:                8109,
		MaxConn:                12000,
		LogPath:                "./log",
		LogName:                "server.log",
		MaxLogNum:              10,
		MaxFileSize:            100,
		LogFileUnit:            logger.KB,
		LogLevel:               logger.ERROR,
		SetToConsole:           true,
		LogFileType:            1,
		PoolSize:               10,
		IsUsePool:              true,
		MaxWorkerLen:           1024 * 2,
		MaxSendChanLen:         1024,
		FrameSpeed:             30,
		OnConnectioned:         func(fconn iface.Iconnection) {},
		OnClosed:               func(fconn iface.Iconnection) {},
		OnClusterConnectioned:  func(fconn iface.Iconnection) {},
		OnClusterClosed:        func(fconn iface.Iconnection) {},
		OnClusterCConnectioned: func(fconn iface.Iclient) {},
		OnClusterCClosed:       func(fconn iface.Iclient) {},
	}
	//读取用户自定义配置
	data, err := ioutil.ReadFile("conf/server.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, &GlobalObject)
	if err != nil {
		panic(err)
	}
}
