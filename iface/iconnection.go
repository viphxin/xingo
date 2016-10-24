package iface

import (
	"github.com/golang/protobuf/proto"
	"net"
)

type Iconnection interface {
	Start()
	Stop()
	GetConnection() *net.TCPConn
	GetSessionId() uint32
	Send(uint32, proto.Message) error
	SendBuff(uint32, proto.Message) error
	RemoteAddr() net.Addr
	LostConnection()
	GetProperty(string) (interface{}, error)
	SetProperty(string, interface{})
	RemoveProperty(string)
}
