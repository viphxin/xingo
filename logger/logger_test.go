package logger

import (
	"runtime"
	"strconv"
	"testing"
	"time"
)

func runlog(i int) {
	Debug("Debug>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Info("Info>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Warn("Warn>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Error("Error>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
	Fatal("Fatal>>>>>>>>>>>>>>>>>>>>>>>>>" + strconv.Itoa(i))
}

func Test(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())

	//指定是否控制台打印，默认为true
	SetConsole(true)
	//指定日志文件备份方式为文件大小的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	//第三个参数为备份文件最大数量
	//第四个参数为备份文件大小
	//第五个参数为文件大小的单位
	//logger.SetRollingFile("d:/logtest", "test.log", 10, 5, logger.KB)

	//指定日志文件备份方式为日期的方式
	//第一个参数为日志文件存放目录
	//第二个参数为日志文件命名
	SetRollingDaily("./logtest", "test.log")

	//指定日志级别  ALL，DEBUG，INFO，WARN，ERROR，FATAL，OFF 级别由低到高
	//一般习惯是测试阶段为debug，生成环境为info以上
	SetLevel(ERROR)

	for i := 10000; i > 0; i-- {
		go runlog(i)
		time.Sleep(1000 * time.Millisecond)
	}
	time.Sleep(15 * time.Second)
}
