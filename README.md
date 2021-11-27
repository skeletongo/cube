# cube

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
