package iface

type ICommandInterpreter interface {
	AddCommand(ICommand)
	Excute(string) string
	IsQuitCmd(string) bool
}

