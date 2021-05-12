package znet

import "github.com/impact-eintr/Zinx/ziface"

// 实现router时 先嵌入BaseRouter基类 type Router struct {BaseRouter}
/*

// 在处理conn业务之前的钩子方法Hook
func (br *BaseRouter) PreHandle(request ziface.IRequest) {}

// 在处理conn业务的主方法
func (br *BaseRouter) Handle(request ziface.IRequest) {}

// 在处理comm也完全之后的钩子方法Hook
func (br *BaseRouter) PostHandle(request ziface.IRequest) {}
*/
type BaseRouter struct{}

// 在处理conn业务之前的钩子方法Hook
func (br *BaseRouter) PreHandle(request ziface.IRequest) {}

// 在处理conn业务的主方法
func (br *BaseRouter) Handle(request ziface.IRequest) {}

// 在处理comm也完全之后的钩子方法Hook
func (br *BaseRouter) PostHandle(request ziface.IRequest) {}
