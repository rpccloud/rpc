// Package server ...
package server

import (
	"crypto/tls"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

// Server ...
type Server struct {
	config        *ServerConfig
	listeners     []*listener
	processor     *rpc.Processor
	sessionServer *SessionServer
	streamHub     *rpc.StreamHub
	mountServices []*rpc.ServiceMeta
	sync.Mutex
}

// NewServer ...
func NewServer(config *ServerConfig) *Server {
	if config == nil {
		config = GetDefaultServerConfig()
	} else {
		config = config.clone()
	}

	return &Server{
		config:        config,
		listeners:     make([]*listener, 0),
		processor:     nil,
		sessionServer: nil,
		streamHub:     nil,
		mountServices: make([]*rpc.ServiceMeta, 0),
	}
}

// Listen ...
func (p *Server) Listen(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *Server {
	p.Lock()
	defer p.Unlock()

	if p.streamHub == nil {
		p.listeners = append(p.listeners, &listener{
			isDebug:   false,
			network:   network,
			addr:      addr,
			tlsConfig: tlsConfig,
		})
	} else {
		p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	}

	return p
}

// ListenWithDebug ...
func (p *Server) ListenWithDebug(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *Server {
	p.Lock()
	defer p.Unlock()

	if p.streamHub == nil {
		p.listeners = append(p.listeners, &listener{
			isDebug:   true,
			network:   network,
			addr:      addr,
			tlsConfig: tlsConfig,
		})
	} else {
		p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	}

	return p
}

// AddService ...
func (p *Server) AddService(
	name string,
	service *rpc.Service,
	data rpc.Map,
) *Server {
	p.Lock()
	defer p.Unlock()

	if p.streamHub == nil {
		p.mountServices = append(p.mountServices, rpc.NewServiceMeta(
			name,
			service,
			base.GetFileLine(1),
			data,
		))
	} else {
		p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	}

	return p
}

// BuildReplyCache ...
func (p *Server) BuildReplyCache() *base.Error {
	p.Lock()
	defer p.Unlock()

	_, file, _, _ := runtime.Caller(1)
	buildDir := path.Join(path.Dir(file))

	processor := rpc.NewProcessor(
		1,
		64,
		64,
		1024,
		nil,
		time.Second,
		p.mountServices,
		rpc.NewTestStreamReceiver(),
	)
	defer processor.Close()

	return processor.BuildCache(
		"cache",
		path.Join(buildDir, "cache", "rpc_action_cache.go"),
	)
}

// Open ...
func (p *Server) Open() bool {
	source := base.GetFileLine(1)

	ret := func() bool {
		p.Lock()
		defer p.Unlock()

		if p.streamHub != nil {
			p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
				base.ErrServerAlreadyRunning.AddDebug(source),
			))
			return false
		}

		processor := (*rpc.Processor)(nil)
		sessionServer := (*SessionServer)(nil)

		streamHub := rpc.NewStreamHub(
			p.config.isLogToScreen,
			p.config.logFile,
			p.config.logLevel,
			rpc.StreamHubCallback{
				OnRPCRequestStream: func(stream *rpc.Stream) {
					processor.PutStream(stream)
				},
				OnRPCResponseOKStream: func(stream *rpc.Stream) {
					sessionServer.OutStream(stream)
				},
				OnRPCResponseErrorStream: func(stream *rpc.Stream) {
					sessionServer.OutStream(stream)
				},
				OnRPCBoardCastStream: func(stream *rpc.Stream) {
					sessionServer.OutStream(stream)
				},
				OnSystemErrorReportStream: func(
					sessionID uint64,
					err *base.Error,
				) {
					// ignore
				},
			},
		)

		processor = rpc.NewProcessor(
			p.config.numOfThreads,
			p.config.maxNodeDepth,
			p.config.maxCallDepth,
			p.config.threadBufferSize,
			p.config.actionCache,
			p.config.closeTimeout,
			p.mountServices,
			streamHub,
		)

		if processor == nil {
			streamHub.Close()
			return false
		}

		sessionServer = NewSessionServer(
			p.listeners,
			p.config.session,
			streamHub,
		)

		p.streamHub = streamHub
		p.processor = processor
		p.sessionServer = sessionServer

		return true
	}()

	if ret {
		p.sessionServer.Open()
	}

	return ret
}

// IsRunning ...
func (p *Server) IsRunning() bool {
	p.Lock()
	defer p.Unlock()

	return p.streamHub != nil
}

// Close ...
func (p *Server) Close() bool {
	p.Lock()
	defer p.Unlock()

	if p.streamHub == nil {
		return false
	}

	if p.sessionServer != nil {
		p.sessionServer.Close()
		p.sessionServer = nil
	}

	if p.processor != nil {
		p.processor.Close()
		p.processor = nil
	}

	p.streamHub.Close()
	p.streamHub = nil
	return true
}
