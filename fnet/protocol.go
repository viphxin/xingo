package fnet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"io"
	"time"
)

const (
	HEADLEN       = 8
	MaxPacketSize = 1024 * 1024
)

var (
	packageTooBig = errors.New("Too many data to receive!!")
)

type PkgData struct {
	Len   uint32
	MsgId uint32
	Data  []byte
}

type PkgAll struct {
	Pdata *PkgData
	Fconn iface.Iconnection
}

type Protocol struct {
	SendBuffChan chan []byte
	ExtSendChan  chan bool
}

func (this *Protocol) OnConnectionMade(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d connected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnConnectioned(fconn)
}

func (this *Protocol) OnConnectionLost(fconn iface.Iconnection) {
	this.ExtSendChan <- true
	logger.Info(fmt.Sprintf("client ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClosed(fconn)
	close(this.ExtSendChan)
	close(this.SendBuffChan)
}

func (this *Protocol) Unpack(headdata []byte) (head *PkgData, err error) {
	headbuf := bytes.NewReader(headdata)

	head = &PkgData{}

	// 读取Len
	if err = binary.Read(headbuf, binary.LittleEndian, &head.Len); err != nil {
		return nil, err
	}

	// 读取MsgId
	if err = binary.Read(headbuf, binary.LittleEndian, &head.MsgId); err != nil {
		return nil, err
	}

	// 封包太大
	if head.Len > MaxPacketSize {
		return nil, packageTooBig
	}

	return head, nil
}

func (this *Protocol) Pack(msgId uint32, data proto.Message) (out []byte, err error) {
	outbuff := bytes.NewBuffer([]byte{})
	// 进行编码
	dataBytes := []byte{}
	if data != nil {
		dataBytes, err = proto.Marshal(data)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("marshaling error:  %s", err))
	}
	// 写Len
	if err = binary.Write(outbuff, binary.LittleEndian, uint32(len(dataBytes))); err != nil {
		return
	}
	// 写MsgId
	if err = binary.Write(outbuff, binary.LittleEndian, msgId); err != nil {
		return
	}

	//all pkg data
	if err = binary.Write(outbuff, binary.LittleEndian, dataBytes); err != nil {
		return
	}

	out = outbuff.Bytes()
	return

}

func (this *Protocol) SendBuff(msgId uint32, data proto.Message) error {
	allData, err := this.Pack(msgId, data)
	if err == nil {

		// 发送超时
		select {
		case <-time.After(time.Second * 2):
			logger.Error(fmt.Sprintf("send error: timeout %d", msgId))
			return errors.New(fmt.Sprintf("send error: timeout %d", msgId))
		case this.SendBuffChan <- allData:
			return nil
		}
	} else {
		logger.Error(fmt.Sprintf("pack data error.reason: %s", err))
	}
	return err
}

/*
send massage to socket imedirtly
*/
func (this *Protocol) Send(fconn iface.Iconnection, msgId uint32, data proto.Message) error {
	allData, err := this.Pack(msgId, data)
	if err != nil {
		logger.Error(fmt.Sprintf("pack data error.reason: %s", err))
		return err
	}

	if _, err = (*fconn.GetConnection()).Write(allData); err != nil {
		logger.Error(fmt.Sprintf("send data error.reason: %s", err))
		return err
	}
	return nil
}

func (this *Protocol) StartReadThread(fconn iface.Iconnection) {
	logger.Info("start receive data from socket...")
	for {
		//read per head data
		headdata := make([]byte, HEADLEN)

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Info("step1 step1step1 step1step1step1step1")
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := this.Unpack(headdata)
		if err != nil {
			logger.Info("step2 step1step1 step1step1step1step1")
			fconn.Stop()
			return
		}
		//data
		if pkgHead.Len > 0 {
			pkgHead.Data = make([]byte, pkgHead.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkgHead.Data); err != nil {
				logger.Info("step3 step1step1 step1step1step1step1")
				fconn.Stop()
				return
			}
		}

		logger.Debug(fmt.Sprintf("msg id :%d, data len: %d", pkgHead.MsgId, pkgHead.Len))
		MsgHandleObj.DoMsg(&PkgAll{
			Pdata: pkgHead,
			Fconn: fconn,
		})
	}
}

func (this *Protocol) StartWriteThread(fconn iface.Iconnection) {
	go func() {
		logger.Info("start send data from channel...")
		for {
			select {
			case <-this.ExtSendChan:
				logger.Info("send thread exit successful!!!!")
				return
			case data := <-this.SendBuffChan:
				//send
				if _, err := (*fconn.GetConnection()).Write(data); err != nil {
					logger.Info("send data error exit...")
					return
				}
			}
		}
	}()
}
