package iface

type Idatapack interface {
	GetHeadLen() int32
	Unpack([]byte) (interface{}, error)
	Pack(uint32, interface{}) ([]byte, error)
}
