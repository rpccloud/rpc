package router

import (
	"crypto/tls"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

// Client ...
type Client struct {
	id         *base.GlobalID
	addr       string
	tlsConfig  *tls.Config
	slot       *Slot
	orcManager *base.ORCManager
}

// NewClient ...
func NewClient(
	addr string,
	tlsConfig *tls.Config,
	streamReceiver rpc.IStreamReceiver,
) (*Client, *base.Error) {
	if streamReceiver == nil {
		panic("streamReceiver is nil")
	}

	id := base.NewGlobalID()

	if id == nil {
		streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrRouterIDInvalid,
		))
		return nil, base.ErrRouterIDInvalid
	}

	ret := &Client{
		id:        id,
		addr:      addr,
		tlsConfig: tlsConfig,
		slot: NewSlot(&ConnectMeta{
			addr:      addr,
			tlsConfig: tlsConfig,
			id:        id,
		}, streamReceiver),
		orcManager: base.NewORCManager(),
	}

	ret.orcManager.Open(func() bool {
		return true
	})

	return ret, nil
}

// SendStream ...
func (p *Client) SendStream(s *rpc.Stream) {
	p.slot.SendStream(s)
}

// Close ...
func (p *Client) Close() bool {
	return p.orcManager.Close(func() {
		p.slot.Close()
	}, func() {
		p.id.Close()
		p.id = nil
	})
}
