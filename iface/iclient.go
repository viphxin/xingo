package iface

import "net"

type Iclient interface {
	Start()
	Stop(bool)
	GetConnection() *net.TCPConn
	Send([]byte) error
	GetProperty(string) (interface{}, error)
	SetProperty(string, interface{})
	RemoveProperty(string)
}
