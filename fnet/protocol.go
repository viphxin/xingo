package fnet

import (
	"errors"
	"fmt"
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"github.com/viphxin/xingo/utils"
	"io"
	"time"
)

const (
	MaxPacketSize = 1024 * 1024
)

var (
	packageTooBig = errors.New("Too many data to received!!")
)

type PkgAll struct {
	Pdata *PkgData
	Fconn iface.Iconnection
}

type Protocol struct {
	msghandle  *MsgHandle
	pbdatapack *PBDataPack
}

func NewProtocol() *Protocol {
	return &Protocol{
		msghandle:  NewMsgHandle(),
		pbdatapack: NewPBDataPack(),
	}
}

func (this *Protocol) GetMsgHandle() iface.Imsghandle {
	return this.msghandle
}
func (this *Protocol) GetDataPack() iface.Idatapack {
	return this.pbdatapack
}

func (this *Protocol) AddRpcRouter(router interface{}) {
	this.msghandle.AddRouter(router)
}

func (this *Protocol) InitWorker(poolsize int32) {
	this.msghandle.StartWorkerLoop(int(poolsize))
}

func (this *Protocol) OnConnectionMade(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d connected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnConnectioned(fconn)
	//加频率控制
	this.SetFrequencyControl(fconn)
}

func (this *Protocol) SetFrequencyControl(fconn iface.Iconnection) {
	fc0, fc1 := utils.GlobalObject.GetFrequency()
	if fc1 == "h" {
		fconn.SetProperty("xingo_fc", 0)
		fconn.SetProperty("xingo_fc0", fc0)
		fconn.SetProperty("xingo_fc1", time.Now().UnixNano()*1e6+int64(3600*1e3))
	} else if fc1 == "m" {
		fconn.SetProperty("xingo_fc", 0)
		fconn.SetProperty("xingo_fc0", fc0)
		fconn.SetProperty("xingo_fc1", time.Now().UnixNano()*1e6+int64(60*1e3))
	} else if fc1 == "s" {
		fconn.SetProperty("xingo_fc", 0)
		fconn.SetProperty("xingo_fc0", fc0)
		fconn.SetProperty("xingo_fc1", time.Now().UnixNano()*1e6+int64(1e3))
	}
}

func (this *Protocol) DoFrequencyControl(fconn iface.Iconnection) error {
	xingo_fc1, err := fconn.GetProperty("xingo_fc1")
	if err != nil {
		//没有频率控制
		return nil
	} else {
		if time.Now().UnixNano()*1e6 >= xingo_fc1.(int64) {
			//init
			this.SetFrequencyControl(fconn)
		} else {
			xingo_fc, _ := fconn.GetProperty("xingo_fc")
			xingo_fc0, _ := fconn.GetProperty("xingo_fc0")
			xingo_fc_int := xingo_fc.(int) + 1
			xingo_fc0_int := xingo_fc0.(int)
			if xingo_fc_int >= xingo_fc0_int {
				//trigger
				return errors.New(fmt.Sprintf("received package exceed limit: %s", utils.GlobalObject.FrequencyControl))
			} else {
				fconn.SetProperty("xingo_fc", xingo_fc_int)
			}
		}
		return nil
	}
}

func (this *Protocol) OnConnectionLost(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	utils.GlobalObject.OnClosed(fconn)
}

func (this *Protocol) StartReadThread(fconn iface.Iconnection) {
	logger.Info("start receive data from socket...")
	for {
		//频率控制
		err := this.DoFrequencyControl(fconn)
		if err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		//read per head data
		headdata := make([]byte, this.pbdatapack.GetHeadLen())

		if _, err := io.ReadFull(fconn.GetConnection(), headdata); err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		pkgHead, err := this.pbdatapack.Unpack(headdata)
		if err != nil {
			logger.Error(err)
			fconn.Stop()
			return
		}
		//data
		pkg := pkgHead.(*PkgData)
		if pkg.Len > 0 {
			pkg.Data = make([]byte, pkg.Len)
			if _, err := io.ReadFull(fconn.GetConnection(), pkg.Data); err != nil {
				logger.Error(err)
				fconn.Stop()
				return
			}
		}

		logger.Debug(fmt.Sprintf("msg id :%d, data len: %d", pkg.MsgId, pkg.Len))
		if utils.GlobalObject.PoolSize > 0 {
			this.msghandle.DeliverToMsgQueue(&PkgAll{
				Pdata: pkg,
				Fconn: fconn,
			})
		} else {
			this.msghandle.DoMsgFromGoRoutine(&PkgAll{
				Pdata: pkg,
				Fconn: fconn,
			})
		}

	}
}
