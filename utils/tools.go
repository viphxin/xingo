package utils

import (
	"fmt"
	"github.com/viphxin/xingo/logger"
	"net/http"
	"reflect"
	"runtime/debug"
	"time"
	"github.com/viphxin/xingo/iface"
)

func HttpRequestWrap(uri string, router iface.IHttpRouter) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.Info("===================http server panic recover===============")
				debug.PrintStack()
			}
		}()
		st := time.Now()
		logger.Debug("User-Agent: ", request.Header["User-Agent"])
		router.PreHandle(response, request)
		router.Handle(response, request)
		router.AfterHandle(response, request)
		logger.Debug(fmt.Sprintf("%s cost total time: %f ms", uri, float64(time.Now().Sub(st).Nanoseconds())/1e6))
	}
}

func ReSettingLog() {
	// --------------------------------------------init log start
	logger.SetConsole(GlobalObject.SetToConsole)
	if GlobalObject.LogFileType == logger.ROLLINGFILE {
		logger.SetRollingFile(GlobalObject.LogPath, GlobalObject.LogName,
			GlobalObject.MaxLogNum, GlobalObject.MaxFileSize, GlobalObject.LogFileUnit)
	} else {
		logger.SetRollingDaily(GlobalObject.LogPath, GlobalObject.LogName)
		logger.SetLevel(GlobalObject.LogLevel)
	}
	// --------------------------------------------init log end
}

func XingoTry(f reflect.Value, args []reflect.Value, handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			logger.Info("-------------panic recover---------------")
			if handler != nil {
				handler(err)
			}
		}
	}()
	f.Call(args)
}