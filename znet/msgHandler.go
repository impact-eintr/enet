package znet

import (
	"fmt"
	"log"

	"github.com/impact-eintr/Zinx/utils"
	"github.com/impact-eintr/Zinx/ziface"
)

// 消息管理实现
type MsgHandle struct {
	// 存放每个MsgID所对应的处理方法(之后是否可以编写路由组)
	Apis map[uint32]ziface.IRouter
	// 负责Worker取任务的消息队列
	TaskQueue []chan ziface.IRequest
	// Worker工作池中Worker的数量
	WorkerPoolSize uint32
}

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		TaskQueue:      make([]chan ziface.IRequest, utils.GlobalConf.WorkerPoolSize),
		WorkerPoolSize: utils.GlobalConf.WorkerPoolSize,
	}
}

// 调度/执行对应的Router消息处理方法
func (h *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	// 1. 从Request中找到msgID
	handler, ok := h.Apis[request.GetMsgID()]
	if !ok {
		fmt.Printf("api msg[%v] is NOT FOUND! Need Register", request.GetMsgID())
	}

	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)

}

// 为消息添加具体的处理逻辑
func (h *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	if _, ok := h.Apis[msgId]; ok {
		log.Println("repeat api, migId = ", msgId)
		return
	}

	h.Apis[msgId] = router
	fmt.Println("Add Router [MsgId] Succ")

}

// 启动一个Worker工作池 开启动作只能发生一次
func (h *MsgHandle) StartWorkerPool() {
	for i := 0; i < int(h.WorkerPoolSize); i++ {
		// 给当前的Worker的TaskQueue开辟空间
		h.TaskQueue[i] = make(chan ziface.IRequest, utils.GlobalConf.TaskQueueSize)
		go h.startOneWorfer(i, h.TaskQueue[i])

	}

}

// 启动一个Worker工作流程
func (h *MsgHandle) startOneWorfer(workerId int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerId, "is started")
	// 不断地阻塞等待对应的消息队列的消息
	for request := range taskQueue {
		h.DoMsgHandler(request)
	}
	//for {
	//	select {
	//	case request := <-taskQueue:
	//		h.DoMsgHandler(request)
	//	}
	//}
}

// 发送消息给消息队列
func (h *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	//根据ConnID来分配当前的连接应该由哪个worker负责处理
	//轮询的平均分配法则
	connID := request.GetConnection().GetConnID()
	workerID := connID % h.WorkerPoolSize
	fmt.Printf("conn[%v]---msg[%v] ---> worker[%v]\n", connID, request.GetMsgID(), workerID)
	h.TaskQueue[workerID] <- request
}
