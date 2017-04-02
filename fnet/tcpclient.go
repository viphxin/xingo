package fnet

import (
	"errors"
	"fmt"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"net"
	"sync"
	"time"
)

const (
	MAX_RETRY      = 1024 //父节点掉线最大重连次数
	RETRY_INTERVAL = 60   //重连间隔60s
)

type TcpClient struct {
	conn          *net.TCPConn
	addr          *net.TCPAddr
	protoc        iface.IClientProtocol
	PropertyBag   map[string]interface{}
	reconnCB      func(iface.Iclient)
	maxRetry      int
	retryInterval int
	sendtagGuard  sync.RWMutex
	propertyLock  sync.RWMutex
}

func NewReConnTcpClient(ip string, port int, protoc iface.IClientProtocol, maxRetry int,
	retryInterval int, reconnCB func(iface.Iclient)) *TcpClient {
	client := NewTcpClient(ip, port, protoc)
	client.maxRetry = maxRetry
	client.retryInterval = retryInterval
	client.reconnCB = reconnCB
	return client
}

func NewTcpClient(ip string, port int, protoc iface.IClientProtocol) *TcpClient {
	addr := &net.TCPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
		Zone: "",
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err == nil {
		client := &TcpClient{
			conn:        conn,
			addr:        addr,
			protoc:      protoc,
			PropertyBag: make(map[string]interface{}, 0),
		}
		go client.protoc.OnConnectionMade(client)
		return client
	} else {
		panic(err)
	}

}

func (this *TcpClient) Start() {
	go this.protoc.StartReadThread(this)
}

func (this *TcpClient) Stop(isforce bool) {
	if this.maxRetry == 0 || isforce {
		this.protoc.OnConnectionLost(this)
	} else {
		//retry
		if this.ReConnection() {
			//顺序很重要，先把读数据用的goroutine开启
			this.Start()
			if this.reconnCB != nil {
				this.reconnCB(this)
			}
		}
	}
}

func (this *TcpClient) ReConnection() bool {
	logger.Info("reconnection ...")
	for i := 1; i <= this.maxRetry; i++ {
		logger.Info("retry time ", i)
		conn, err := net.DialTCP("tcp", nil, this.addr)
		if err == nil {
			this.conn = conn
			return true
		} else {
			d, err := time.ParseDuration(fmt.Sprintf("%ds", this.retryInterval))
			if err != nil {
				time.Sleep(RETRY_INTERVAL * time.Second)
			} else {
				time.Sleep(d)
			}
		}
	}
	return false
}

func (this *TcpClient) Send(data []byte) error {
	this.sendtagGuard.Lock()
	defer this.sendtagGuard.Unlock()

	if _, err := this.conn.Write(data); err != nil {
		logger.Error(fmt.Sprintf("rpc client send data error.reason: %s", err))
		return err
	}
	return nil
}

func (this *TcpClient) GetConnection() *net.TCPConn {
	return this.conn
}

func (this *TcpClient) GetProperty(key string) (interface{}, error) {
	this.propertyLock.RLock()
	defer this.propertyLock.RUnlock()

	value, ok := this.PropertyBag[key]
	if ok {
		return value, nil
	} else {
		return nil, errors.New("no property in connection")
	}
}

func (this *TcpClient) SetProperty(key string, value interface{}) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	this.PropertyBag[key] = value
}

func (this *TcpClient) RemoveProperty(key string) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	delete(this.PropertyBag, key)
}
