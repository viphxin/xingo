package fserver

import (
	"github.com/viphxin/xingo/fnet"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"net"
	_ "time"
)

func init() {
	utils.GlobalObject.Protoc = &fnet.Protocol{
		SendBuffChan: make(chan []byte, utils.GlobalObject.MaxSendChanLen),
		ExtSendChan:  make(chan bool, 1),
	}
	// --------------------------------------------init log start
	logger.SetConsole(utils.GlobalObject.SetToConsole)
	// logger.SetRollingFile(utils.GlobalObject.LogPath, utils.GlobalObject.LogName,
	// utils.GlobalObject.MaxLogNum, utils.GlobalObject.MaxFileSize, utils.GlobalObject.LogFileUnit)
	logger.SetRollingDaily(utils.GlobalObject.LogPath, utils.GlobalObject.LogName)
	logger.SetLevel(utils.GlobalObject.LogLevel)
	// --------------------------------------------init log end
}

type Server struct {
	Port    int
	MaxConn int
	GenNum  uint32
}

func NewServer() *Server {
	s := &Server{
		Port:    utils.GlobalObject.TcpPort,
		MaxConn: utils.GlobalObject.MaxConn,
	}
	return s
}

func (this *Server) handleConnection(conn *net.TCPConn) {
	this.GenNum += 1
	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)
	// conn.SetDeadline(time.Now().Add(time.Minute * 2))
	fconn := fnet.NewConnection(conn, this.GenNum, &fnet.Protocol{
		SendBuffChan: make(chan []byte, utils.GlobalObject.MaxSendChanLen),
		ExtSendChan:  make(chan bool, 1),
	})
	fconn.Start()
}

func (this *Server) Start() {
	go func() {
		//init workpool
		fnet.MsgHandleObj.InitWorkerPool(10)
		ln, err := net.ListenTCP("tcp", &net.TCPAddr{
			Port: this.Port,
		})
		if err != nil {
			logger.Error(err)
		}
		logger.Info("start server...")
		for {
			conn, err := ln.AcceptTCP()
			if err != nil {
				logger.Error(err)
			}
			go this.handleConnection(conn)
		}
	}()
}

func (this *Server) Stop() {
	logger.Info("Stop Server!!!")
}

func (this *Server) AddRouter(router interface{}) {
	logger.Info("AddRouter")
	fnet.MsgHandleObj.AddRouter(router)
}
