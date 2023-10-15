package rpc

import (
	"context"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/transmeta"
	"github.com/cloudwego/kitex/pkg/rpcinfo/remoteinfo"
	"github.com/flyerxp/lib/config"
	"github.com/flyerxp/lib/logger"
	"strconv"
)

var ClientTTHeaderHandler remote.MetaHandler = &clientTTHeaderHandler{}

// clientTTHeaderHandler implement remote.MetaHandler
type clientTTHeaderHandler struct{}

// WriteMeta of clientTTHeaderHandler writes headers of TTHeader protocol to transport
func (ch *clientTTHeaderHandler) WriteMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	ri := msg.RPCInfo()
	transInfo := msg.TransInfo()
	logId := ""
	hd := map[uint16]string{
		transmeta.ToService:   ri.To().ServiceName(),
		transmeta.ToMethod:    ri.To().Method(),
		transmeta.MsgType:     strconv.Itoa(int(msg.MessageType())),
		transmeta.LogID:       logId,
		transmeta.FromService: ri.From().ServiceName(),
		transmeta.FromMethod:  ri.From().Method(),
	}
	if config.GetConf().App.Type != "rpc" {
		hd[transmeta.LogID] = logger.GetLogId(ctx)
	} else {
		remoteObj := remoteinfo.AsRemoteInfo(ri.From())
		if remoteObj == nil {
			return ctx, nil
		}
		strInfo := transInfo.TransIntInfo()
		logId = strInfo[transmeta.LogID]
	}
	transInfo.PutTransIntInfo(hd)
	msg.TransInfo().PutTransIntInfo(hd)
	if metainfo.HasMetaInfo(ctx) {
		hds := make(map[string]string)
		metainfo.SaveMetaInfoToMap(ctx, hds)
		transInfo.PutTransStrInfo(hds)
	}
	return ctx, nil
}

// ReadMeta of clientTTHeaderHandler reads headers of TTHeader protocol from transport
func (ch *clientTTHeaderHandler) ReadMeta(ctx context.Context, msg remote.Message) (context.Context, error) {
	return ctx, nil
}
