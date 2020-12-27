package router

import (
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/core"
)

// IRouteReceiver ...
type IRouteReceiver interface {
	ReceiveStreamFromRouter(stream *core.Stream) *base.Error
}

// IRouteSender ...
type IRouteSender interface {
	SendStreamToRouter(stream *core.Stream) *base.Error
}

// IRouter ...
type IRouter interface {
	Plug(receiver IRouteReceiver) IRouteSender
}
