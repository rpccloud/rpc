package adapter

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
)

// NewSyncClientService ...
func NewSyncClientService(adapter *Adapter) base.IORCService {
	switch adapter.network {
	case "tcp4":
		fallthrough
	case "tcp6":
		fallthrough
	case "tcp":
		fallthrough
	case "ws":
		fallthrough
	case "wss":
		return &syncClientService{
			adapter:    adapter,
			conn:       nil,
			orcManager: base.NewORCManager(),
		}
	default:
		adapter.receiver.OnConnError(
			nil,
			base.ErrUnsupportedProtocol.AddDebug(
				fmt.Sprintf("unsupported protocol %s", adapter.network),
			),
		)
		return nil
	}
}

// NewSyncServerService ...
func NewSyncServerService(adapter *Adapter) base.IORCService {
	switch adapter.network {
	case "tcp4":
		fallthrough
	case "tcp6":
		fallthrough
	case "tcp":
		return &syncTCPServerService{
			adapter:    adapter,
			ln:         nil,
			orcManager: base.NewORCManager(),
		}
	case "ws":
		fallthrough
	case "wss":
		return &syncWSServerService{
			adapter:    adapter,
			ln:         nil,
			server:     nil,
			orcManager: base.NewORCManager(),
		}
	default:
		adapter.receiver.OnConnError(
			nil,
			base.ErrUnsupportedProtocol.AddDebug(
				fmt.Sprintf("unsupported protocol %s", adapter.network),
			),
		)
		return nil
	}
}

func runIConn(conn IConn) {
	conn.OnOpen()
	for {
		if !conn.OnReadReady() {
			break
		}
	}
	conn.OnClose()
}

func runNetConnOnServers(adapter *Adapter, conn net.Conn) {
	go func() {
		syncConn := NewServerSyncConn(conn, adapter.rBufSize, adapter.wBufSize)
		syncConn.SetNext(
			NewStreamConn(adapter.isDebug, syncConn, adapter.receiver),
		)

		runIConn(syncConn)
		syncConn.Close()
	}()
}

// -----------------------------------------------------------------------------
// syncTCPServerService
// -----------------------------------------------------------------------------
type syncTCPServerService struct {
	adapter    *Adapter
	ln         net.Listener
	orcManager *base.ORCManager
}

// Open ...
func (p *syncTCPServerService) Open() bool {
	return p.orcManager.Open(func() bool {
		e := error(nil)
		adapter := p.adapter
		if p.adapter.tlsConfig == nil {
			p.ln, e = net.Listen(adapter.network, adapter.addr)
		} else {
			p.ln, e = tls.Listen(
				adapter.network,
				adapter.addr,
				adapter.tlsConfig,
			)
		}

		if e != nil {
			adapter.receiver.OnConnError(
				nil,
				base.ErrSyncTCPServerServiceListen.AddDebug(e.Error()),
			)
			return false
		}

		return true
	})
}

// Run ...
func (p *syncTCPServerService) Run() {
	p.orcManager.Run(func(isRunning func() bool) {
		adapter := p.adapter
		for isRunning() {
			conn, e := p.ln.Accept()
			if e != nil {
				isCloseErr := p.orcManager.IsClosing() &&
					strings.HasSuffix(e.Error(), base.ErrNetClosingSuffix)

				if !isCloseErr {
					adapter.receiver.OnConnError(
						nil,
						base.ErrSyncTCPServerServiceAccept.AddDebug(e.Error()),
					)
					base.WaitWhileRunning(
						base.TimeNow().UnixNano(),
						isRunning,
						500*time.Millisecond,
					)
				}
			} else {
				runNetConnOnServers(adapter, conn)
			}
		}
	})
}

// Close ...
func (p *syncTCPServerService) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.ln.Close(); e != nil {
			p.adapter.receiver.OnConnError(
				nil,
				base.ErrSyncTCPServerServiceClose.AddDebug(e.Error()),
			)
		}
	}, func() {
		p.ln = nil
	})
}

// -----------------------------------------------------------------------------
// syncWSServerService
// -----------------------------------------------------------------------------
type syncWSServerService struct {
	adapter    *Adapter
	ln         net.Listener
	server     *http.Server
	orcManager *base.ORCManager
}

// Open ...
func (p *syncWSServerService) Open() bool {
	return p.orcManager.Open(func() bool {
		adapter := p.adapter
		path := adapter.path
		if path == "" {
			path = "/"
		}

		mux := http.NewServeMux()

		for k, v := range p.adapter.fileMap {
			mux.Handle(k, http.FileServer(http.Dir(v)))
		}

		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			conn, _, _, e := ws.UpgradeHTTP(r, w)

			if e != nil {
				adapter.receiver.OnConnError(
					nil,
					base.ErrSyncWSServerServiceUpgrade.AddDebug(e.Error()),
				)
			} else {
				runNetConnOnServers(adapter, newSyncWSServerConn(conn))
			}
		})

		p.server = &http.Server{
			Addr:    adapter.addr,
			Handler: mux,
		}

		e := error(nil)

		if adapter.tlsConfig == nil {
			p.ln, e = net.Listen("tcp", adapter.addr)
		} else {
			p.ln, e = tls.Listen("tcp", adapter.addr, adapter.tlsConfig)
		}

		if e != nil {
			adapter.receiver.OnConnError(
				nil,
				base.ErrSyncWSServerServiceListen.AddDebug(e.Error()),
			)
			return false
		}

		return true
	})
}

// Run ...
func (p *syncWSServerService) Run() {
	p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			startNS := base.TimeNow().UnixNano()
			if e := p.server.Serve(p.ln); e != nil {
				if e != http.ErrServerClosed {
					p.adapter.receiver.OnConnError(
						nil,
						base.ErrSyncWSServerServiceServe.AddDebug(e.Error()),
					)
				}
			}
			base.WaitWhileRunning(startNS, isRunning, time.Second)
		}
	})
}

// Close ...
func (p *syncWSServerService) Close() bool {
	return p.orcManager.Close(func() {
		if e := p.server.Close(); e != nil {
			p.adapter.receiver.OnConnError(
				nil,
				base.ErrSyncWSServerServiceClose.AddDebug(e.Error()),
			)
		}

		if e := p.ln.Close(); e != nil {
			if !strings.HasSuffix(e.Error(), base.ErrNetClosingSuffix) {
				p.adapter.receiver.OnConnError(
					nil,
					base.ErrSyncWSServerServiceClose.AddDebug(e.Error()),
				)
			}
		}
	}, func() {
		p.server = nil
		p.ln = nil
	})
}

// -----------------------------------------------------------------------------
// syncClientService
// -----------------------------------------------------------------------------
type syncClientService struct {
	adapter    *Adapter
	conn       *SyncConn
	orcManager *base.ORCManager
	mu         sync.Mutex
}

func (p *syncClientService) openConn() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	var e error
	var conn net.Conn
	var wsRawConn net.Conn

	adapter := p.adapter
	switch adapter.network {
	case "tcp4":
		fallthrough
	case "tcp6":
		fallthrough
	case "tcp":
		if adapter.tlsConfig == nil {
			conn, e = net.Dial(adapter.network, adapter.addr)
		} else {
			conn, e = tls.Dial(adapter.network, adapter.addr, adapter.tlsConfig)
		}
	case "ws":
		fallthrough
	case "wss":
		dialer := &ws.Dialer{TLSConfig: adapter.tlsConfig}
		u := url.URL{Scheme: adapter.network, Host: adapter.addr, Path: "/"}
		wsRawConn, _, _, e = dialer.Dial(context.Background(), u.String())
		conn = newSyncWSClientConn(wsRawConn)
	default:
		adapter.receiver.OnConnError(
			nil,
			base.ErrUnsupportedProtocol.AddDebug(
				fmt.Sprintf("unsupported protocol %s", adapter.network),
			),
		)
		return false
	}

	if e != nil {
		adapter.receiver.OnConnError(
			nil,
			base.ErrSyncClientServiceDial.AddDebug(e.Error()),
		)
		return false
	}

	p.conn = NewClientSyncConn(conn, adapter.rBufSize, adapter.wBufSize)
	p.conn.SetNext(NewStreamConn(p.adapter.isDebug, p.conn, p.adapter.receiver))
	return true
}

func (p *syncClientService) closeConn() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn := p.conn; conn != nil {
		conn.Close()
	}
}

// Open ...
func (p *syncClientService) Open() bool {
	return p.orcManager.Open(func() bool {
		return true
	})
}

// Run ...
func (p *syncClientService) Run() {
	p.orcManager.Run(func(isRunning func() bool) {
		for isRunning() {
			startNS := base.TimeNow().UnixNano()

			if p.openConn() {
				runIConn(p.conn)
				p.closeConn()
			}

			base.WaitWhileRunning(
				startNS,
				isRunning,
				3*time.Second,
			)
		}
	})
}

// Close ...
func (p *syncClientService) Close() bool {
	return p.orcManager.Close(func() {
		p.closeConn()
	}, func() {
		p.conn = nil
	})
}
