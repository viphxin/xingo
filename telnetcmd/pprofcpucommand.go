package telnetcmd

import (
	"bytes"
	"fmt"
	"runtime/pprof"
)

type PprofCpuCommand struct {
	buffer *bytes.Buffer
	profilingBuffer *bytes.Buffer
}

func NewPprofCpuCommand() *PprofCpuCommand{
	return &PprofCpuCommand{new(bytes.Buffer), new(bytes.Buffer)}
}
func (this *PprofCpuCommand)Name()string{
	return "pprofcpu"
}

func (this *PprofCpuCommand)Help()string{
	return fmt.Sprintf("pprofcpu:\r\n" +
		"----------- start: 开始收集服务cpu占用信息\r\n" +
		"----------- stop:  结束数据收集\r\n" +
		"----------- profiling: 分析(goroutine, heap, thread, block)")
}

func (this *PprofCpuCommand)Run(args []string) string{
	if len(args) == 0{
		return this.Help()
	}else{
		switch args[0] {
		case "start":
			pprof.StopCPUProfile()
			this.buffer.Reset()
			err := pprof.StartCPUProfile(this.buffer)
			if err != nil{
				return fmt.Sprintf("pprofcpu start error: %s.", err)
			}else{
				return "pprofcpu start successful. please wait."
			}
		case "stop":
			pprof.StopCPUProfile()
			return this.buffer.String()
		case "profiling":
			var p  *pprof.Profile
			switch args[1] {
			case "goroutine":
				p = pprof.Lookup("goroutine")
			case "heap":
				p = pprof.Lookup("heap")
			case "thread":
				p = pprof.Lookup("threadcreate")
			case "block":
				p = pprof.Lookup("block")
			default:
				return this.Help()
			}
			this.profilingBuffer.Reset()
			err := p.WriteTo(this.profilingBuffer, 1)
			if err != nil {
				return fmt.Sprintf("pprofcpu profiling error: %s.", err)
			}else{
				return this.profilingBuffer.String() + "\r\n"
			}
		default:
			return "not found command."

		}
	}
}
