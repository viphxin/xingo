package iface

type Imsghandle interface {
	DeliverToMsgQueue(interface{})
	DoMsgFromGoRoutine(interface{})
	AddRouter(string, IRouter)
	AddRpcRouter(string, IRpcRouter)
	StartWorkerLoop(int)
}
