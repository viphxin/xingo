package cluster

import (
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"fmt"
	"time"
	"bufio"
	"strings"
	"github.com/viphxin/xingo/utils"
)
/*
debug tool protocol
*/

type TelnetProtocol struct {}

func NewTelnetProtocol() *TelnetProtocol {
	if utils.GlobalObject.CmdInterpreter == nil{
		utils.GlobalObject.CmdInterpreter = NewCommandInterpreter()
	}
	return &TelnetProtocol{}
}

func (this *TelnetProtocol) GetMsgHandle() iface.Imsghandle {
	return nil
}
func (this *TelnetProtocol) GetDataPack() iface.Idatapack {
	return nil
}

func (this *TelnetProtocol) AddRpcRouter(router interface{}) {

}

func (this *TelnetProtocol) InitWorker(poolsize int32) {

}

func (this *TelnetProtocol)isWriteListIP(ip string) bool{
	for _, wip := range utils.GlobalObject.WriteList{
		if strings.EqualFold(ip, wip){
			return true
		}
	}
	return false
}

func (this *TelnetProtocol)getConnectionAddr(fconn iface.Iconnection)[]string{
	return strings.Split(fconn.RemoteAddr().String(), ":")
}

func (this *TelnetProtocol) OnConnectionMade(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d connected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
	addr := this.getConnectionAddr(fconn)
	if !this.isWriteListIP(addr[0]){
		logger.Error("invald IP: ", addr[0])
		fconn.LostConnection()
	}
}

func (this *TelnetProtocol) OnConnectionLost(fconn iface.Iconnection) {
	logger.Info(fmt.Sprintf("client ID: %d disconnected. IP Address: %s", fconn.GetSessionId(), fconn.RemoteAddr()))
}

func (this *TelnetProtocol) StartReadThread(fconn iface.Iconnection) {
	logger.Info("start receive data from telnet socket...")
	fconn.GetConnection().Write([]byte(fmt.Sprintf("-------welcome to xingo telnet tool(node: %s)---------\r\n", utils.GlobalObject.Name)))
	for {
		if err := fconn.GetConnection().SetReadDeadline(time.Now().Add(time.Minute*3)); err != nil {
			logger.Error("telnet connection SetReadDeadline error: ", err)
			fconn.LostConnection()
			break
		}
		line, err := bufio.NewReader(fconn.GetConnection()).ReadString('\n')
		if err != nil {
			logger.Error("telnet connection read line error: ", err)
			fconn.Stop()
			break
		}
		line = strings.TrimSuffix(line[:len(line)-1], "\r")
		logger.Info(fmt.Sprintf("xingo telnet tool received: %s. ip: %s", line, this.getConnectionAddr(fconn)[0]))
		if utils.GlobalObject.CmdInterpreter.IsQuitCmd(line){
			logger.Error("telnet exit ")
			fconn.LostConnection()
			break
		}else{
			ack := utils.GlobalObject.CmdInterpreter.Excute(line)
			fconn.GetConnection().Write([]byte(ack))
		}
		if err := fconn.GetConnection().SetReadDeadline(time.Time{}); err != nil {
			logger.Error("telnet connection SetReadDeadline error: ", err)
			fconn.LostConnection()
			break
		}

	}
}

