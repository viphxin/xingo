# xingo
高性能golang网络库，游戏开发脚手架

默认通信协议如下（支持自定义协议处理部分代码，支持灵活的重载协议部分代码）：

Len   uint32 数据Data部分长度
MsgId uint32 消息号
Data  []byte 数据
消息默认通过google 的protobuf进行序列化

服务器全局配置对象为GlobalObject，支持的配置选项及默认值如下：

  |TcpPort:        |8109,//服务器监听端口
  |MaxConn:        |12000,//支持最大链接数
  |LogPath:        |"./log",//日志文件路径
  |LogName:        "server.log",//日志文件名
  |MaxLogNum:      |10,//最大日志数
  |MaxFileSize:    |100,//per日志文件大小
  |LogFileUnit:    |logger.KB,//日志文件大小对应单位
  |LogLevel:       |logger.ERROR,//日志级别
  |SetToConsole:   |true,//是否输出到console
  |PoolSize:       10,//api接口工作线程数量
  |MaxWorkerLen:   |1024 * 2,//任务缓冲池大小
  |MaxSendChanLen: |1024,//发送队列从缓冲池
  |FrameSpeed:     |30,//未使用
  |OnConnectioned: |func(fconn iface.Iconnection) {},//链接建立事件回调
  |OnClosed:       |func(fconn iface.Iconnection) {},//链接断开事件回调
  
  如何使用？
  只需要一步，添加消息路由：
  
  //add api ---------------start
	FightingRouterObj := &api.FightingRouter{}
	s.AddRouter(FightingRouterObj)
	//add api ---------------end
  xingo会自动注册FightingRouter中的方法处理对应消息
  例如：msgId =1 则会寻找FightingRouter中的Func_1的方法从进行处理
  具体使用请参考项目（也是xingo的demo）：
  https://github.com/viphxin/fighting
  
