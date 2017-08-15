package telnetcmd

import (
	"fmt"
	"github.com/viphxin/xingo/clusterserver"
	"strconv"
)

type CloseServerCommand struct {
}

func NewCloseServerCommand() *CloseServerCommand{
	return &CloseServerCommand{}
}
func (this *CloseServerCommand)Name()string{
	return "closeserver"
}

func (this *CloseServerCommand)Help()string{
	return fmt.Sprintf("closeserver:\r\n" +
		"----------- all delay: 延迟delay秒时间关闭所有子节点\r\n" +
		"----------- node name delay:  延迟delay秒时间关闭指定节点\r\n")
}

func (this *CloseServerCommand)Run(args []string) string{
	if len(args) == 0{
		return this.Help()
	}else{
		switch args[0] {
		case "all":
			for _, child := range clusterserver.GlobalMaster.Childs.GetChilds() {
				if len(args) > 1{
					if v, err := strconv.ParseInt(args[1], 10, 64); err == nil{
						child.CallChildNotForResult("CloseServer", int(v))
					}else{
						child.CallChildNotForResult("CloseServer", int(0))
					}
				}else{
					child.CallChildNotForResult("CloseServer", int(0))
				}
			}
		default:
			child, err := clusterserver.GlobalMaster.Childs.GetChild(args[0])
			if err != nil{
				return fmt.Sprintf("no sush node: %s.", args[0])
			}else{
				if len(args) > 1{
					if v, err := strconv.ParseInt(args[1], 10, 64); err == nil{
						child.CallChildNotForResult("CloseServer", int(v))
					}else{
						child.CallChildNotForResult("CloseServer", int(0))
					}
				}else{
					child.CallChildNotForResult("CloseServer", int(0))
				}
			}
		}
	}
	return "OK"
}

type ReloadCfgCommand struct {
}

func NewReloadCfgCommand() *ReloadCfgCommand{
	return &ReloadCfgCommand{}
}
func (this *ReloadCfgCommand)Name()string{
	return "reloadcfg"
}

func (this *ReloadCfgCommand)Help()string{
	return fmt.Sprintf("reloadcfg:\r\n" +
		"----------- all delay: 延迟delay秒时间重新加载所有节点的配置文件\r\n" +
		"----------- node name delay:  延迟delay秒时间重新加载指定节点\r\n")
}

func (this *ReloadCfgCommand)Run(args []string) string{
	if len(args) == 0{
		return this.Help()
	}else{
		switch args[0] {
		case "all":
			for _, child := range clusterserver.GlobalMaster.Childs.GetChilds() {
				if len(args) > 1{
					if v, err := strconv.ParseInt(args[1], 10, 64); err == nil{
						child.CallChildNotForResult("ReloadConfig", int(v))
					}else{
						child.CallChildNotForResult("ReloadConfig", int(0))
					}
				}else{
					child.CallChildNotForResult("ReloadConfig", int(0))
				}
			}
		default:
			child, err := clusterserver.GlobalMaster.Childs.GetChild(args[0])
			if err != nil{
				return fmt.Sprintf("no sush node: %s.", args[0])
			}else{
				if len(args) > 1{
					if v, err := strconv.ParseInt(args[1], 10, 64); err == nil{
						child.CallChildNotForResult("ReloadConfig", int(v))
					}else{
						child.CallChildNotForResult("ReloadConfig", int(0))
					}
				}else{
					child.CallChildNotForResult("ReloadConfig", int(0))
				}
			}
		}
	}
	return "OK"
}
