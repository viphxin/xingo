package iface

type IWriter interface {
	Send([]byte) error
	GetProperty(string) (interface{}, error)
	SetProperty(string, interface{})
	RemoveProperty(string)
}
