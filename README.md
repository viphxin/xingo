# xingo_cluster
==================
xingo golang游戏开发交流群：535378240<br>
文档地址: http://www.w3cschool.cn/xingo/
```text
    xingo_cluster是免费、开源、可定制、可扩展、节点支持“热更新”的高性能分布式游戏服务器开发框架，采用golang语言开发，天
生携带高并发场景的处理基因，继承了golang语言本身的各种优点，开发简单易上手并且功能强大。它主要实现了高性能的异步网络库
xingo，分布式节点间的高性能rpc通信，日志管理等，可以节省大量游戏开发时间，让游戏开发人员可以将主要精力放到游戏玩法和游戏逻
辑上。真正实现了修改配置文件就可以搭建自定义的分布式游戏服务器架构。

优势特点：
1) 开发效率高
2) 支持自定义的分布式架构，方便横向扩展节点，理论上只要有足够的物理机器，没有承载上限
3) 支持自定义通信协议
4) 分布式节点自动发现
5) worker pool工作线程池
```
示例配置:<br>
```json
{
    "master":{"host": "192.168.2.225","rootport":9999},
    "servers":{
        "gate2":{"host": "192.168.2.225", "rootport":10000,"name":"gate2", "module": "gate", "log": "gate2.log"},
        "gate1":{"host": "192.168.2.225", "rootport":10001,"name":"gate1", "module": "gate", "log": "gate1.log"},
        "net1":{"host": "192.168.2.225", "netport":11009,"name":"net1","remotes":["gate2", "gate1"], 
                    "module": "net", "log": "net.log"},
        "net2":{"host": "192.168.2.225", "netport":11010,"name":"net2","remotes":["gate2", "gate1"], 
                    "module": "net", "log": "net.log"},
        "net3":{"host": "192.168.2.225", "netport":11011,"name":"net3","remotes":["gate2", "gate1"], 
                    "module": "net", "log": "net.log"},
        "net4":{"host": "192.168.2.225", "netport":11012,"name":"net4","remotes":["gate2", "gate1"], 
                    "module": "net", "log": "net.log"},
        "admin":{"host": "192.168.2.225", "remotes":["gate2", "gate1"], "name":"admin", "module": "admin", 
            "http": [8888, "/static"]},
        "game1":{"host": "192.168.2.225", "remotes":["gate2", "gate1"], "name":"game1", "module": "game"}
    }
}
```
架构图：
![alt text](https://git.oschina.net/viphxin/xingo_cluster/raw/master/conf/xingo_cluster_架构.png)


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
  LogFileType:    1,//日志切割方式1 按天切割 2按文件大小切割
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
  具体使用请参考项目：<br>
  帧同步服务器: https://github.com/viphxin/fighting<br>
  mmo demo: https://git.oschina.net/viphxin/xingo_demo<br>
  xingo_cluster demo: https://github.com/viphxin/xingo_cluster
  
