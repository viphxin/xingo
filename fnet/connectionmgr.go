package fnet

import (
	"errors"
	"github.com/viphxin/xingo/logger"
	"sync"
)

type ConnectionMsg struct {
	connections map[uint32]*Connection
	conMrgLock  sync.RWMutex
}

func (this *ConnectionMsg) Add(conn *Connection) {
	this.conMrgLock.Lock()
	defer this.conMrgLock.Unlock()
	this.connections[conn.SessionId] = conn
	logger.Info(len(this.connections))
}

func (this *ConnectionMsg) Remove(conn *Connection) error {
	this.conMrgLock.Lock()
	defer this.conMrgLock.Unlock()
	_, ok := this.connections[conn.SessionId]
	if ok {
		delete(this.connections, conn.SessionId)
		logger.Info(len(this.connections))
		return nil
	} else {
		return errors.New("not found!!")
	}

}

func (this *ConnectionMsg) Get(sid uint32) (*Connection, error) {
	v, ok := this.connections[sid]
	if ok {
		delete(this.connections, sid)
		return v, nil
	} else {
		return nil, errors.New("not found!!")
	}
}

func (this *ConnectionMsg) Len() int{
	return len(this.connections)
}

var ConnectionManager *ConnectionMsg

func init() {
	ConnectionManager = &ConnectionMsg{
		connections: make(map[uint32]*Connection),
	}
}
