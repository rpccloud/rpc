// Package adapter ...
package adapter

import (
	"crypto/tls"
	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
	"net"
)

// IConn ...
type IConn interface {
	OnOpen()
	OnClose()
	OnError(err *base.Error)
	OnReadReady() bool
	OnWriteReady() bool
	OnReadBytes(b []byte)
	OnFillWrite(b []byte) int

	SetNext(conn IConn)
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	Close()
}

// IReceiver ...
type IReceiver interface {
	OnConnOpen(streamConn *StreamConn)
	OnConnClose(streamConn *StreamConn)
	OnConnReadStream(streamConn *StreamConn, stream *rpc.Stream)
	OnConnError(streamConn *StreamConn, err *base.Error)
}

// Adapter ...
type Adapter struct {
	isDebug    bool
	isClient   bool
	network    string
	addr       string
	tlsConfig  *tls.Config
	rBufSize   int
	wBufSize   int
	receiver   IReceiver
	service    base.IORCService
	orcManager *base.ORCManager
}

// NewClientAdapter ...
func NewClientAdapter(
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *Adapter {
	return &Adapter{
		isDebug:    false,
		isClient:   true,
		network:    network,
		addr:       addr,
		tlsConfig:  tlsConfig,
		rBufSize:   rBufSize,
		wBufSize:   wBufSize,
		receiver:   receiver,
		service:    nil,
		orcManager: base.NewORCManager(),
	}
}

// NewServerAdapter ...
func NewServerAdapter(
	isDebug bool,
	network string,
	addr string,
	tlsConfig *tls.Config,
	rBufSize int,
	wBufSize int,
	receiver IReceiver,
) *Adapter {
	return &Adapter{
		isDebug:    isDebug,
		isClient:   false,
		network:    network,
		addr:       addr,
		tlsConfig:  tlsConfig,
		rBufSize:   rBufSize,
		wBufSize:   wBufSize,
		receiver:   receiver,
		service:    nil,
		orcManager: base.NewORCManager(),
	}
}

// Open ...
func (p *Adapter) Open() bool {
	return p.orcManager.Open(func() bool {
		if p.isClient {
			p.service = NewSyncClientService(p)
		} else {
			p.service = NewSyncServerService(p)
		}

		if p.service == nil {
			return false
		}

		return p.service.Open()
	})
}

// Run ...
func (p *Adapter) Run() bool {
	return p.orcManager.Run(func(isRunning func() bool) bool {
		return p.service.Run()
	})
}

// Close ...
func (p *Adapter) Close() bool {
	return p.orcManager.Close(func() bool {
		return p.service.Close()
	}, func() {
		p.service = nil
	})
}
