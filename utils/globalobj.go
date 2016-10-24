package utils

import (
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
)

type GlobalObj struct {
	OnConnectioned func(fconn iface.Iconnection)
	OnClosed       func(fconn iface.Iconnection)
	Protoc         interface{}
	TcpPort        int
	MaxConn        int
	//log
	LogPath        string
	LogName        string
	MaxLogNum      int32
	MaxFileSize    int64
	LogFileUnit    logger.UNIT
	LogLevel       logger.LEVEL
	SetToConsole   bool
	PoolSize       int32
	MaxWorkerLen   int32
	MaxSendChanLen int32
	FrameSpeed     uint8
}

var GlobalObject *GlobalObj

func init() {
	GlobalObject = &GlobalObj{
		TcpPort:        8109,
		MaxConn:        12000,
		LogPath:        "./log",
		LogName:        "server.log",
		MaxLogNum:      10,
		MaxFileSize:    100,
		LogFileUnit:    logger.KB,
		LogLevel:       logger.ERROR,
		SetToConsole:   true,
		PoolSize:       10,
		MaxWorkerLen:   1024 * 2,
		MaxSendChanLen: 1024,
		FrameSpeed:     30,
		OnConnectioned: func(fconn iface.Iconnection) {},
		OnClosed:       func(fconn iface.Iconnection) {},
	}
}
