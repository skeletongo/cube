# cube

### feature
* 服务发现
* 支持rpc
* 完善线程监控

#### 简介
游戏开发框架，提供基础功能，如：网络通信、日志、配置、定时任务、线程监控、模块管理等

#### 日志说明  
日志默认使用三方库 https://github.com/sirupsen/logrus  
修改或扩展日志功能只要修改或替换掉这个三方库中默认的标准日志对象就可以了  
通过 logrus.StandardLogger() 获取三方库中默认的日志对象

#### 代码说明  
* object: 基础节点，单线程模型，包含一个消息队列及定时器，在单线程中串行处理消息队列中的所有消息及定时任务
* module: 自定义功能模块  
    * network: 提供网络服务，支持tcp,websocket，过滤器network.Filter，中间件network.Middle  
* timer: 创建延迟函数及定时任务  
* g: 多线程支持
* statsviz: 查看程序运行时的工具库 https://github.com/arl/statsviz

#### 配置文件
``` 
# 模块配置
module:
  Options:
    Interval: 100 # 定时器间隔，单位毫秒
# 网络配置
network:
  Endian: false # 字节序，默认为小端序，true表示大端序
  IsJson: false # 修改默认编码方式为json,否则是encoding/gob，encoding/gob是go语言特有的编码方式
  LenMsgLen: 2 # 封包时应用层数据长度所占用的字节数
  MinMsgLen: 1 # 封包时应用层数据最短字节数
  MaxMsgLen: 4096 # 封包时应用层数据最大字节数
  Services:
    - Area: 1 # 服务区域
      Type: 1 # 服务类型
      ID: 1 # 服务ID
      CertFile: # 证书文件地址
      KeyFile: # 秘钥文件地址
      Name: CubeTcpServer # 服务名称
      Protocol: tcp # 服务协议，tcp/ws/wss
      Ip: 127.0.0.1 # 服务IP
      OutIp: 127.0.0.1 # 服务外网IP
      Port: 8888 # 服务端口
      MaxConnNum: 1000 # 最大连接数
      MaxRecv: 4096 # 消息接收队列长度
      MaxSend: 4096 # 消息发送队列长度
      Linger: 0 # TCP连接关闭时，延迟关闭的时间，单位秒，0立即关闭
      KeepAlive: false # 是否启用TCP的KeepAlive，默认为false
      KeepAlivePeriod: 0 # TCP的KeepAlive周期，单位秒，0表示使用系统默认值
      ReadBufferSize: 0 # 读缓冲区大小，0表示使用系统默认值
      WriteBufferSize: 0 # 写缓冲区大小，0表示使用系统默认值
      ReadTimeout: 0 # 读超时时间，单位秒，0表示不设置超时时间
      WriteTimeout: 0 # 写超时时间，单位秒，0表示不设置超时时间
      FilterChain: ["auth"] # 使用的过滤器名称及顺序
      MiddleChain: [] # 使用的中间件名称及顺序
    - Area: 1
      Type: 1
      ID: 2
      Name: CubeWSServer
      CertFile: ''
      KeyFile: ''
      Protocol: ws
      Ip: 127.0.0.1
      Port: 8889
      Path: /
      ReadBufferSize: 0
      WriteBufferSize: 0
      HTTPTimeout: 0
      ReadTimeout: 0
      WriteTimeout: 0
      FilterChain: []
      MiddleChain: []
    - Area: 1
      Type: 2
      ID: 1
      Name: CubeTcpClient
      IsClient: true # 是否为客户端
      IsAutoReconnect: true # 是否自动重连
      ReconnectInterval: 3 # 重连间隔，单位秒
      Protocol: tcp
      Ip: 127.0.0.1
      Port: 8888
      Linger: 0
      KeepAlive: false
      KeepAlivePeriod: 0
      ReadBufferSize: 0
      WriteBufferSize: 0
      ReadTimeout: 0
      WriteTimeout: 0
      FilterChain: []
      MiddleChain: []
    - Area: 1
      Type: 2
      ID: 2
      Name: CubeWSClient
      IsClient: true
      IsAutoReconnect: true
      Protocol: ws
      Ip: 127.0.0.1
      Port: 8889
      Path: /
      ClientNum: 1
      ReadBufferSize: 0
      WriteBufferSize: 0
      HTTPTimeout: 0
      ReadTimeout: 0
      WriteTimeout: 0
      FilterChain: []
      MiddleChain: []
# github.com/arl/statsviz
statsviz:
  IsOpen: true # 是否开启
  Addr: ':6060' # 地址
```