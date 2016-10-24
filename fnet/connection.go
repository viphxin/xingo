package fnet

import (
	"errors"
	"github.com/golang/protobuf/proto"
	_ "github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"net"
	"sync"
	"time"
)

type Connection struct {
	Conn         *net.TCPConn
	isClosed     bool
	SessionId    uint32
	Protoc       *Protocol
	PropertyBag  map[string]interface{}
	sendtagGuard sync.RWMutex
}

func NewConnection(conn *net.TCPConn, sessionId uint32, protoc *Protocol) *Connection {
	fconn := &Connection{
		Conn:        conn,
		isClosed:    false,
		SessionId:   sessionId,
		Protoc:      protoc,
		PropertyBag: make(map[string]interface{}),
	}
	//set  connection time
	fconn.SetProperty("ctime", time.Since(time.Now()))
	return fconn
}

func (this *Connection) Start() {
	//add to connectionmsg
	ConnectionManager.Add(this)
	this.Protoc.OnConnectionMade(this)
	this.Protoc.StartWriteThread(this)
	this.Protoc.StartReadThread(this)
}

func (this *Connection) Stop() {
	// 防止将Send放在go内造成的多线程冲突问题
	this.sendtagGuard.Lock()
	defer this.sendtagGuard.Unlock()

	this.isClosed = true
	this.Protoc.OnConnectionLost(this)
	//remove to connectionmsg
	ConnectionManager.Remove(this)
}

func (this *Connection) GetConnection() *net.TCPConn {
	return this.Conn
}

func (this *Connection) GetSessionId() uint32 {
	return this.SessionId
}

func (this *Connection) GetProtoc() *Protocol {
	return this.Protoc
}

func (this *Connection) GetProperty(key string) (interface{}, error) {
	value, ok := this.PropertyBag[key]
	if ok {
		return value, nil
	} else {
		return nil, errors.New("no property in connection")
	}
}

func (this *Connection) SetProperty(key string, value interface{}) {
	this.PropertyBag[key] = value
}

func (this *Connection) RemoveProperty(key string) {
	delete(this.PropertyBag, key)
}

func (this *Connection) Send(msgId uint32, data proto.Message) error {
	if !this.isClosed {
		return this.Protoc.Send(this, msgId, data)
	} else {
		return errors.New("connection closed")
	}
}

func (this *Connection) SendBuff(msgId uint32, data proto.Message) error {
	// 防止将Send放在go内造成的多线程冲突问题
	this.sendtagGuard.Lock()
	defer this.sendtagGuard.Unlock()

	if !this.isClosed {
		return this.Protoc.SendBuff(msgId, data)
	} else {
		return errors.New("connection closed")
	}

}

func (this *Connection) RemoteAddr() net.Addr {
	return (*this.Conn).RemoteAddr()
}

func (this *Connection) LostConnection() {
	(*this.Conn).Close()
	this.isClosed = true
	logger.Info("LostConnection==============!!!!!!!!!!!!!!!!!!!!!!!!!")
}
