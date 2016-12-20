# xingo
高性能golang网络库，游戏开发脚手架<br>

默认通信协议如下（支持自定义协议处理部分代码，支持灵活的重载协议部分代码）：<br>

Len   uint32 数据Data部分长度<br>
MsgId uint32 消息号<br>
Data  []byte 数据<br>
消息默认通过google 的protobuf进行序列化<br>

服务器全局配置对象为GlobalObject，支持的配置选项及默认值如下：<br>
  TcpPort:        8109,//服务器监听端口<br>
  MaxConn:        12000,//支持最大链接数<br>
  LogPath:        "./log",//日志文件路径<br>
  LogName:        "server.log",//日志文件名<br>
  MaxLogNum:      10,//最大日志数<br>
  MaxFileSize:    100,//per日志文件大小<br>
  LogFileUnit:    logger.KB,//日志文件大小对应单位<br>
  LogLevel:       logger.ERROR,//日志级别<br>
  SetToConsole:   true,//是否输出到console<br>
  PoolSize:       10,//api接口工作线程数量<br>
  IsUsePool:      true,//是否使用worker pool false 每个请求开启单独的协程处理<br>
  MaxWorkerLen:   1024 * 2,//任务缓冲池大小<br>
  MaxSendChanLen: 1024,//发送队列从缓冲池<br>
  FrameSpeed:     30,//未使用<br>
  OnConnectioned: func(fconn iface.Iconnection) {},//链接建立事件回调<br>
  OnClosed:       func(fconn iface.Iconnection) {},//链接断开事件回调<br>
  
  如何使用？<br>
  只需要一步，添加消息路由：<br>
  s := fserver.NewServer()<br>
  //add api ---------------start<br>
	FightingRouterObj := &api.FightingRouter{}<br>
	s.AddRouter(FightingRouterObj)<br>
	//add api ---------------end<br>
  xingo会自动注册FightingRouter中的方法处理对应消息<br>
  例如：msgId =1 则会寻找FightingRouter中的Func_1的方法从进行处理<br>
  具体使用请参考项目（也是xingo的demo（帧同步服务器））：<br>
  https://github.com/viphxin/fighting<br>
  mmo demo: https://git.oschina.net/viphxin/xingo_demo
  
