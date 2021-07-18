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

const (
	defaultMaxNumOfThreads  = 1024 * 1024
	defaultThreadsPerCPU    = 16384
	defaultThreadBufferSize = 2048
	defaultCloseTimeout     = 5 * time.Second
	defaultMaxNodeDepth     = 128
	defaultMaxCallDepth     = 128
)

var fnNumCPU = runtime.NumCPU

// Server ...
type Server struct {
	isRunning        bool
	processor        *rpc.Processor
	sessionServer    *SessionServer
	numOfThreads     int
	maxNodeDepth     int16
	maxCallDepth     int16
	threadBufferSize uint32
	actionCache      rpc.ActionCache
	closeTimeout     time.Duration
	mountServices    []*rpc.ServiceMeta
	logReceiver      rpc.IStreamReceiver
	sync.Mutex
}

// NewServer ...
func NewServer(logReceiver rpc.IStreamReceiver) *Server {
	if logReceiver == nil {
		logReceiver = rpc.NewLogToScreenErrorStreamReceiver("Server")
	}

	ret := &Server{
		isRunning:        false,
		processor:        nil,
		sessionServer:    nil,
		numOfThreads:     fnNumCPU() * defaultThreadsPerCPU,
		maxNodeDepth:     defaultMaxNodeDepth,
		maxCallDepth:     defaultMaxCallDepth,
		threadBufferSize: defaultThreadBufferSize,
		actionCache:      nil,
		closeTimeout:     defaultCloseTimeout,
		mountServices:    make([]*rpc.ServiceMeta, 0),
		logReceiver:      logReceiver,
	}

	if ret.numOfThreads > defaultMaxNumOfThreads {
		ret.numOfThreads = defaultMaxNumOfThreads
	}

	ret.sessionServer = NewSessionServer(GetDefaultSessionConfig(), ret)

	return ret
}

// Listen ...
func (p *Server) Listen(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *Server {
	p.sessionServer.Listen(network, addr, tlsConfig)
	return p
}

// ListenWithDebug ...
func (p *Server) ListenWithDebug(
	network string,
	addr string,
	tlsConfig *tls.Config,
) *Server {
	p.sessionServer.ListenWithDebug(network, addr, tlsConfig)
	return p
}

// SetNumOfThreads ...
func (p *Server) SetNumOfThreads(numOfThreads int) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else if numOfThreads <= 0 {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrNumOfThreadsIsWrong.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.numOfThreads = numOfThreads
	}

	return p
}

// SetThreadBufferSize ...
func (p *Server) SetThreadBufferSize(threadBufferSize uint32) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else if threadBufferSize <= 0 {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrThreadBufferSizeIsWrong.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.threadBufferSize = threadBufferSize
	}

	return p
}

// SetActionCache ...
func (p *Server) SetActionCache(actionCache rpc.ActionCache) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.actionCache = actionCache
	}

	return p
}

// SetLogReceiver ...
func (p *Server) SetLogReceiver(logReceiver rpc.IStreamReceiver) *Server {
	p.Lock()
	defer p.Unlock()

	if p.isRunning {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.logReceiver = logReceiver
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

	if p.isRunning {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerAlreadyRunning.AddDebug(base.GetFileLine(1)),
		))
	} else {
		p.mountServices = append(p.mountServices, rpc.NewServiceMeta(
			name,
			service,
			base.GetFileLine(1),
			data,
		))
	}

	return p
}

// BuildReplyCache ...
func (p *Server) BuildReplyCache() *Server {
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

	if err := processor.BuildCache(
		"cache",
		path.Join(buildDir, "cache", "rpc_action_cache.go"),
	); err != nil {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(err))
	}

	return p
}

// OnReceiveStream ...
func (p *Server) OnReceiveStream(stream *rpc.Stream) {
	if stream != nil {
		switch stream.GetKind() {
		case rpc.StreamKindRPCRequest:
			p.processor.PutStream(stream)
		case rpc.StreamKindRPCResponseOK:
			fallthrough
		case rpc.StreamKindRPCResponseError:
			fallthrough
		case rpc.StreamKindRPCBoardCast:
			p.sessionServer.OutStream(stream)
		default:
			if stream.GetKind() == rpc.StreamKindSystemErrorReport {
				p.logReceiver.OnReceiveStream(stream)
			} else {
				stream.Release()
			}
		}
	}
}

// Open ...
func (p *Server) Open() bool {
	source := base.GetFileLine(1)

	ret := func() bool {
		p.Lock()
		defer p.Unlock()

		if p.isRunning {
			p.OnReceiveStream(rpc.MakeSystemErrorStream(
				base.ErrServerAlreadyRunning.AddDebug(source),
			))
			return false
		} else if processor := rpc.NewProcessor(
			p.numOfThreads,
			p.maxNodeDepth,
			p.maxCallDepth,
			p.threadBufferSize,
			p.actionCache,
			p.closeTimeout,
			p.mountServices,
			p,
		); processor == nil {
			return false
		} else {
			p.isRunning = true
			p.processor = processor
			return true
		}
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

	return p.isRunning
}

// Close ...
func (p *Server) Close() bool {
	p.Lock()
	defer p.Unlock()

	if !p.isRunning {
		p.OnReceiveStream(rpc.MakeSystemErrorStream(
			base.ErrServerNotRunning.AddDebug(base.GetFileLine(1)),
		))
		return false
	}

	p.sessionServer.Close()
	p.processor.Close()
	p.processor = nil
	p.isRunning = false
	return true
}
