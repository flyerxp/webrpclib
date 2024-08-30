package rpc

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/transmeta"
	"github.com/flyerxp/lib/v2/logger"
)

var ServerTTHeaderHandler remote.MetaHandler = &serverTTHeaderHandler{}

// serverTTHeaderHandler implement remote.MetaHandler
type serverTTHeaderHandler struct{}

// WriteMeta of serverTTHeaderHandler writes headers of TTHeader protocol to transport
func (ch *serverTTHeaderHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}

// ReadMeta of serverTTHeaderHandler reads headers of TTHeader protocol from transport
func (ch *serverTTHeaderHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	transInfo := msg.TransInfo().TransIntInfo()
	if id, ok := transInfo[transmeta.LogID]; ok {
		ctx = logger.GetContext(ctx, id)
	}
	rpcInfo := msg.RPCInfo()
	if method, ok := transInfo[transmeta.FromMethod]; ok {
		logger.SetRefer(ctx, fmt.Sprintf("%s.%s", rpcInfo.From().ServiceName(), method))
	} else {
		logger.SetRefer(ctx, rpcInfo.From().ServiceName())
	}
	logger.SetUrl(ctx, rpcInfo.To().ServiceName()+"."+rpcInfo.To().Method())
	logger.SetAddr(ctx, rpcInfo.From().Address().String(), serverIp)
	return ctx, nil
}
