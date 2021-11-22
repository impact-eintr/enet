package enet

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/impact-eintr/enet/iface"
)

type MsgHandle struct {
	Apis           map[uint32]iface.IRouter // 存放每个MsgId 所对应的处理方法的map属性
	WorkerPoolSize uint32                   //业务工作Worker池的数量
	TaskQueue      []chan iface.IRequest    //Worker负责取任务的消息队列
}

func NewMsgHandler() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]iface.IRouter),
		WorkerPoolSize: GlobalObject.WorkerPoolSize,
		//一个worker对应一个queue
		TaskQueue: make([]chan iface.IRequest, GlobalObject.WorkerPoolSize)}
}

// 马上以非阻塞的方式处理消息
func (mh *MsgHandle) DoMsgHandler(request iface.IRequest) {
	if request.GetMsgID() == 10 && string(request.GetData()) == "让我访问!!!" {
		log.Panic(request.GetMsgID())
	}
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	// 执行对应处理方法
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router iface.IRouter) {
	// 1 判断当前msg绑定的API处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api , msgId = " + strconv.Itoa(int(msgId)))
	}
	// 2 添加msg与api的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

//启动一个Worker工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan iface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	// 不断的等待队列中的消息
	for {
		select {
		// 有消息则取出队列的Request，并执行绑定的业务方法
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// 启动worker工作池
func (mh *MsgHandle) StartWorkerPool() {
	// 遍历需要启动worker的数量，依此启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个worker被启动
		// 给当前worker对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan iface.IRequest, GlobalObject.MaxWorkerTaskLen)
		// 启动当前Worker，阻塞的等待对应的任务队列是否有消息传递进来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// 将消息交给TaskQueue,由worker进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request iface.IRequest) {
	// 根据ConnID来分配当前的连接应该由哪个worker负责处理
	// 随机数分配原则

	// 得到需要处理此条连接的workerID
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	workerID := r.Intn(int(mh.WorkerPoolSize))
	fmt.Println("Add ConnID=", request.GetConnection().GetConnID(),
		" request msgID=", request.GetMsgID(), "to workerID=", workerID)
	// 将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}
