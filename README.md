# cube

#### 日志说明  
日志默认使用三方库 https://github.com/sirupsen/logrus  
修改或扩展日志功能只要修改或替换掉这个三方库中默认的标准日志对象就可以了  
通过 logrus.StandardLogger() 获取三方库中默认的日志对象

#### 代码说明  
* object: 基础节点，包含一个消息队列，并在自己的协程中串行处理消息队列中的所有消息  
* module: 自定义功能模块  
    * network: 网络服务 tcp,udp,websocket  
* timer: 创建延迟函数及定时任务  
* task: 创建任务，对多线程的支持  
* worker: 相当于一个协程，处理Task任务  
