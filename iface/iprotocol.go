package iface

type IProtocol interface {
	OnConnectionMade(fconn Iconnection)
	OnConnectionLost(fconn Iconnection)
	Unpack(data []byte)
	Pack(hdata interface{}, data interface{}) ([]byte, error)
	StartReadThread(fconn Iconnection)
}
