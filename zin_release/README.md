# Zinx TCP服务器框架

## v0.2
### 简单的连接封装和业务绑定
> 连接的模块
- 方法
    - 启动连接
    - 停止连接
    - 获取当前连接的conn对象(socket)
    - 得到连接ID
    - 得到客户端连接的地址和端口
    - 发送数据的方法
- 属性
    - socket TCP套接字
    - 连接的ID
    - 当前连接的状态
    - 与当前连接绑定的处理业务方法
    - 等待连接被动退出的channel

## v0.3

> 基础router模块

- Ruquest请求封装(将链接和数据绑定在一起)
    - 属性
        - 连接IConnection
        - 请求数据
    - 方法
        - 得到当前连接
        - 得到当前数据
- Router模块
    - 抽象的IRouter
        -  处理业务之前的方法
        -  处理业务的方法
        -  处理业务之后的方法
    - 具体的BaseRouter(作为具体实现的基类)
        -  处理业务之前的方法
        -  处理业务的方法
        -  处理业务之后的方法
- zinx集成router模块
    - Iserver增添路由功能
    - Server类增添Router成员
    - Commection类绑定一个Router成员
    - 在Connection调用 已经注册的Router处理业务

## v0.4
> 增添全局配置

## v0.5

> 消息封装

- 定义一个消息的结构
    - 属性
        - 消息的ID
        - 消息长度
        - 消息的内容
    - 方法
        - Setter
        - Getter
- 将消息封装机制集成到Zinx框架中    
    - 将Message添加到Request中
    - 修改连接读取数据的机制 将之前的单纯读取byte改为拆包读取方式
    - 连接的发包机制 将发送的消息进行打包 再发送

## v0.6

> 消息管理模块

- 属性
    - 集合-消息ID和对应的router的关系 map
- 方法
    - 根据msgID来索引调度路由方法
    - 添加路由方法到map集合中
> 将消息管理机制集成到Zinx框架中
1. 将server模块中的Router属性 替换成MsgHandler属性
2. 将server之前的AddRouter修改成AddRouter--AddRouter(msgId unit32, router ziface.IRouter)
3. 将connection模块Router属性 替换成MsgHandler 修改Connection方法
4. Connection的之前调度Router的业务替换成MsgHandler调度 修改StartReader方法

## v0.7
> Zinx读写分离

![Zinx读写分离](https://img.kancloud.cn/80/28/8028019d6bfce107ebc1bf5a15fd8940_1024x768.jpeg)

1. 添加一个Reader与Write之间通信的channel
2. 添加一个Writer Goroutine
3. Reader由之前直接发送给客户端 改为发送给通信Channel
4. 启动Reader和Writer一同工作


## v0.8
> 消息队列以及多任务

![消息队列](https://img.kancloud.cn/70/6c/706cb06abebcb8c1b7dd22c23d79cf48_1024x768.jpeg)

1. 创建一个消息队列
- MsgHandler消息管理模块
    - 增加属性
        - 消息队列
        - worker工作池的数量
2. 创建多任务worker的工作池并且启动
- 创建一个Worker的工作池并且启动
    - 根据Workerpoolsize的数量去创建Worker
    - 每一个Worker都开启一个协程负载
        - 阻塞等待与当前Worker对应的channel来消息
        - 一旦有消息到来mworker应该处理当前消息对应的业务
3. 将之前的发送消息，全部都改成发送给消息队列和Worker工作池来处理
- 定义一个方法 将消息发送给消息队列工作池
    - 保证每个worker所受的request任务书均衡(平均分配) 
    - 将消息发送给对应的队列 

> 将消息队列机制集成到Zinx框架中
- 开启并调用消息队列
- 将从客户端接收的数据 发送给当前的Worker工作池来处理

## v0.9
> 创建一个链接管理模块 
- 属性
    - 已经创建的Connection集合
    - 针对map的互斥锁
- 方法
    - 添加链接
    - 删除链接
    - 根据链接ID查找对应的链接
    - 总链接个数
    - 清除全部的链接
> 将链接管理模块集成到Zinx框架中

> 给Zinx框架提供 创建链接之后/销毁链接之前 的钩子函数

>

## v1.0

