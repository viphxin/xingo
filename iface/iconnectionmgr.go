package iface

type Iconnectionmgr interface {
	Add(Iconnection)
	Remove(Iconnection) error
	Get(uint32) (Iconnection, error)
	Len() int
}
