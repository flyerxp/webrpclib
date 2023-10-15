package rpc

import (
	"context"
	"github.com/cloudwego/kitex/pkg/remote"
	"net"
)

// InboundRpcHandler is used to process read event.
type InboundRpcHandler struct {
}

var inboundRpc = InboundRpcHandler{}

func (i *InboundRpcHandler) OnActive(ctx context.Context, conn net.Conn) (context.Context, error) {

	return ctx, nil
}
func (i *InboundRpcHandler) OnRead(ctx context.Context, conn net.Conn) (context.Context, error) {
	return ctx, nil
}
func (i *InboundRpcHandler) OnInactive(ctx context.Context, conn net.Conn) context.Context {
	return ctx
}
func (i *InboundRpcHandler) OnMessage(ctx context.Context, args, msg remote.Message) (context.Context, error) {
	return ctx, nil
}
