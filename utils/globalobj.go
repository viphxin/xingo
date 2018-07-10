package utils

import (
	"encoding/json"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"io/ioutil"
	"strconv"
	"strings"
	"os"
	"github.com/viphxin/xingo/timer"
)

type GlobalObj struct {
	TcpServers             map[string]iface.Iserver
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
	Host                   string
	DebugPort              int          //telnet port 用于单机模式
	WriteList              []string     //telnet ip list
	TcpPort                int
	MaxConn                int
	IntraMaxConn           int          //内部服务器最大连接数
	//log
	LogPath          string
	LogName          string
	MaxLogNum        int32
	MaxFileSize      int64
	LogFileUnit      logger.UNIT
	LogLevel         logger.LEVEL
	SetToConsole     bool
	LogFileType      int32
	PoolSize         int32
	MaxWorkerLen     int32
	MaxSendChanLen   int32
	FrameSpeed       uint8
	Name             string
	MaxPacketSize    uint32
	FrequencyControl string //  100/h, 100/m, 100/s
	CmdInterpreter   iface.ICommandInterpreter //xingo debug tool Interpreter
	ProcessSignalChan chan os.Signal
	safeTimerScheduel *timer.SafeTimerScheduel
}

func (this *GlobalObj) GetFrequency() (int, string) {
	fc := strings.Split(this.FrequencyControl, "/")
	if len(fc) != 2 {
		return 0, ""
	} else {
		fc0_int, err := strconv.Atoi(fc[0])
		if err == nil {
			return fc0_int, fc[1]
		} else {
			logger.Error("FrequencyControl params error: ", this.FrequencyControl)
			return 0, ""
		}
	}
}

func (this *GlobalObj)IsThreadSafeMode()bool{
	if this.PoolSize == 1{
		return true
	}else{
		return false
	}
}

func (this *GlobalObj)GetSafeTimer() *timer.SafeTimerScheduel{
	return this.safeTimerScheduel
}

func (this *GlobalObj)Reload(){
	//读取用户自定义配置
	data, err := ioutil.ReadFile("conf/server.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(data, this)
	if err != nil {
		panic(err)
	}else{
		ReSettingLog()
		//init safetimer
		if GlobalObject.safeTimerScheduel == nil && GlobalObject.IsThreadSafeMode(){
			GlobalObject.safeTimerScheduel = timer.NewSafeTimerScheduel()
		}
	}
}

var GlobalObject *GlobalObj

func init() {
	GlobalObject = &GlobalObj{
		TcpServers: make(map[string]iface.Iserver),
		Host:                   "0.0.0.0",
		TcpPort:                8109,
		MaxConn:                12000,
		IntraMaxConn:           100,
		LogPath:                "./log",
		LogName:                "server.log",
		MaxLogNum:              10,
		MaxFileSize:            100,
		LogFileUnit:            logger.KB,
		LogLevel:               logger.DEBUG,
		SetToConsole:           true,
		LogFileType:            1,
		PoolSize:               10,
		MaxWorkerLen:           1024 * 2,
		MaxSendChanLen:         1024,
		FrameSpeed:             30,
		OnConnectioned:         func(fconn iface.Iconnection) {},
		OnClosed:               func(fconn iface.Iconnection) {},
		OnClusterConnectioned:  func(fconn iface.Iconnection) {},
		OnClusterClosed:        func(fconn iface.Iconnection) {},
		OnClusterCConnectioned: func(fconn iface.Iclient) {},
		OnClusterCClosed:       func(fconn iface.Iclient) {},
		ProcessSignalChan:      make(chan os.Signal, 1),
	}
	GlobalObject.Reload()
}
