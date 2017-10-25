package utils

import (
	"fmt"
)

const (
	MAXUINT32 = 4294967295
	DEFAULT_UUID_CNT_CACHE = 512
)

type UUIDGenerator struct {
	Prefix       string
	idGen        uint32
	internalChan chan uint32
}

func NewUUIDGenerator(prefix string) *UUIDGenerator {
	gen := &UUIDGenerator{
		Prefix:       prefix,
		idGen:        0,
		internalChan: make(chan uint32, DEFAULT_UUID_CNT_CACHE),
	}
	gen.startGen()
	return gen
}

func (this *UUIDGenerator) startGen() {
	go func() {
		for {
			if this.idGen+1 > MAXUINT32 {
				this.idGen = 1
			} else {
				this.idGen += 1
			}
			this.internalChan <- this.idGen
		}
	}()
}

func (this *UUIDGenerator) Get() string {
	idgen := <-this.internalChan
	return fmt.Sprintf("%s%d", this.Prefix, idgen)
}

func (this *UUIDGenerator) GetUint32() uint32 {
	return <-this.internalChan
}