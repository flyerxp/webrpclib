package rpc

import (
	"context"
	"fmt"
	"github.com/cloudwego/kitex/pkg/remote"
	"github.com/cloudwego/kitex/pkg/remote/transmeta"
	"github.com/flyerxp/lib/logger"
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
	logger.SetUrl(ctx, rpcInfo.To().ServiceName()+"."+rpcInfo.To().Method())
	logger.SetRefer(ctx, fmt.Sprintf("%s.%s", rpcInfo.From().ServiceName(), rpcInfo.From().Method()))
	logger.SetAddr(ctx, rpcInfo.From().Address().String(), serverIp)
	return ctx, nil
}

/*func encryption(FromVal *reflect.Value, FromType reflect.Type) {
	for i := 0; i < FromType.NumField(); i++ {
		formValField := FromVal.Field(i)
		fieldType := FromType.Field(i).Type.Name()
		if !(fieldType == "string") {
			continue
		}
		oldVal := formValField.String()
		oldLen := len(oldVal)
		if oldLen > 128 {
			formValField.SetString(oldVal[0:128])
		}
		tag := FromType.Field(i).Tag.Get("encryption")
		if tag != "" {
			ArrTag := strings.Split(tag, ",")
			enType := "right"
			if len(ArrTag) > 1 {
				enType = ArrTag[1]
			}
			enLen, err := strconv.Atoi(ArrTag[0])
			if err != nil {
				enLen = 3
			}
			if (oldLen - enLen) < 1 {
				continue
			}
			newVal := ""
			stopChar := len(oldVal) - enLen
			switch enType {
			case "left":
				newVal = strings.Repeat("*", enLen) + oldVal[enLen:]
			case "center":
				stopChar = int(oldLen/2) - int(enLen/2)
				newVal = oldVal[0:stopChar] + strings.Repeat("*", enLen) + oldVal[stopChar+enLen:]
			default:
				newVal = oldVal[0:stopChar] + strings.Repeat("*", enLen)
			}
			formValField.SetString(newVal)
		}
	}
}
*/
