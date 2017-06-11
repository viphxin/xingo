package cluster

import (
	"github.com/viphxin/xingo/iface"
	"github.com/viphxin/xingo/logger"
	"strings"
	"fmt"
)

var (
	QUIT_CMD = [3]string{"quit", "q", "exit"}
)

type CommandInterpreter struct {
	commands map[string]iface.ICommand
}

func NewCommandInterpreter() *CommandInterpreter{
	interpreter := &CommandInterpreter{make(map[string]iface.ICommand)}
	return interpreter
}

func (this *CommandInterpreter)AddCommand(cmd iface.ICommand){
	this.commands[cmd.Name()] = cmd
	logger.Debug("add command ", cmd.Name())
}

func (this *CommandInterpreter)preExcute(rawCmdExp string) string{
	return strings.ToLower(strings.TrimSpace(rawCmdExp))
}

func (this *CommandInterpreter)IsQuitCmd(rawCmdExp string) bool{
	cmdExp := this.preExcute(rawCmdExp)
	for _, cmd := range QUIT_CMD{
		if cmd == cmdExp{
			return true
		}
	}
	return false
}

func (this *CommandInterpreter)help() string{
	helpStr := "有关某个命令的详细信息，请键入 help 命令名"
	for _, v := range this.commands{
		helpStr = fmt.Sprintf("%s\r\n%s", helpStr, v.Help())
	}
	return helpStr
}

func (this *CommandInterpreter)Excute(rawCmdExp string) string{
	defer func()string{
		if err := recover(); err != nil{
			logger.Error("invalid rawCmdExp: ", rawCmdExp)
			return "invalid rawCmdExp: " + rawCmdExp
		}
		return "Unkown ERROR!!!"
	}()
	if rawCmdExp == ""{
		return ""
	}
	rawCmdExps := strings.Split(rawCmdExp, " ")
	if len(rawCmdExps) == 0{
		return ""
	}
	cmdExps := make([]string, 0)
	for _, cmd := range rawCmdExps{
		cmdExps = append(cmdExps, this.preExcute(cmd))
	}

	if command, ok := this.commands[cmdExps[0]]; ok{
		return command.Run(cmdExps[1:])
	}else{
		if cmdExps[0] == "help"{
			return this.help()
		}else{
			return "command not found."
		}
	}
}

