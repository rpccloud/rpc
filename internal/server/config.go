package server

import (
	"crypto/tls"
	"net/http"
	"runtime"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

type listener struct {
	isDebug   bool
	network   string
	addr      string
	path      string
	fileMap   map[string]http.Handler
	tlsConfig *tls.Config
}

// SessionConfig ...
type SessionConfig struct {
	numOfChannels     int
	transLimit        int
	heartbeatInterval time.Duration
	heartbeatTimeout  time.Duration

	serverMaxSessions     int
	serverSessionTimeout  time.Duration
	serverReadBufferSize  int
	serverWriteBufferSize int
	serverCacheTimeout    time.Duration
}

// GetDefaultSessionConfig ...
func GetDefaultSessionConfig() *SessionConfig {
	return &SessionConfig{
		numOfChannels:         32,
		transLimit:            4 * 1024 * 1024,
		heartbeatInterval:     4 * time.Second,
		heartbeatTimeout:      8 * time.Second,
		serverMaxSessions:     10240000,
		serverSessionTimeout:  120 * time.Second,
		serverReadBufferSize:  1200,
		serverWriteBufferSize: 1200,
		serverCacheTimeout:    10 * time.Second,
	}
}

func (p *SessionConfig) SetNumOfChannels(numOfChannels int) *SessionConfig {
	p.numOfChannels = numOfChannels
	return p
}

func (p *SessionConfig) SetTransLimit(transLimit int) *SessionConfig {
	p.transLimit = transLimit
	return p
}

func (p *SessionConfig) SetHeartbeatInterval(
	heartbeatInterval time.Duration,
) *SessionConfig {
	p.heartbeatInterval = heartbeatInterval
	return p
}

func (p *SessionConfig) SetHeartbeatTimeout(
	heartbeatTimeout time.Duration,
) *SessionConfig {
	p.heartbeatTimeout = heartbeatTimeout
	return p
}

func (p *SessionConfig) SetServerMaxSessions(
	serverMaxSessions int,
) *SessionConfig {
	p.serverMaxSessions = serverMaxSessions
	return p
}

func (p *SessionConfig) SetServerSessionTimeout(
	serverSessionTimeout time.Duration,
) *SessionConfig {
	p.serverSessionTimeout = serverSessionTimeout
	return p
}

func (p *SessionConfig) SetServerReadBufferSize(
	serverReadBufferSize int,
) *SessionConfig {
	p.serverReadBufferSize = serverReadBufferSize
	return p
}

func (p *SessionConfig) SetServerWriteBufferSize(
	serverWriteBufferSize int,
) *SessionConfig {
	p.serverWriteBufferSize = serverWriteBufferSize
	return p
}

func (p *SessionConfig) SetServerCacheTimeout(
	serverCacheTimeout time.Duration,
) *SessionConfig {
	p.serverCacheTimeout = serverCacheTimeout
	return p
}

func (p *SessionConfig) clone() *SessionConfig {
	return &SessionConfig{
		numOfChannels:         p.numOfChannels,
		transLimit:            p.transLimit,
		heartbeatInterval:     p.heartbeatInterval,
		heartbeatTimeout:      p.heartbeatTimeout,
		serverMaxSessions:     p.serverMaxSessions,
		serverSessionTimeout:  p.serverSessionTimeout,
		serverReadBufferSize:  p.serverReadBufferSize,
		serverWriteBufferSize: p.serverWriteBufferSize,
		serverCacheTimeout:    p.serverCacheTimeout,
	}
}

type ServerConfig struct {
	logToScreen      bool
	logFile          string
	logLevel         base.ErrorLevel
	numOfThreads     int
	maxNodeDepth     int16
	maxCallDepth     int16
	threadBufferSize uint32
	closeTimeout     time.Duration
	actionCache      rpc.ActionCache
	session          *SessionConfig
}

func GetDefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		logToScreen:      true,
		logFile:          "",
		logLevel:         base.ErrorLogAll,
		numOfThreads:     base.MinInt(runtime.NumCPU(), 64) * 16384,
		maxNodeDepth:     128,
		maxCallDepth:     128,
		threadBufferSize: 2048,
		closeTimeout:     5 * time.Second,
		actionCache:      nil,
		session:          GetDefaultSessionConfig(),
	}
}

func (p *ServerConfig) SetLogToScreen(logToScreen bool) *ServerConfig {
	p.logToScreen = logToScreen
	return p
}

func (p *ServerConfig) SetLogFile(logFile string) *ServerConfig {
	p.logFile = logFile
	return p
}

func (p *ServerConfig) SetLogLevel(logLevel base.ErrorLevel) *ServerConfig {
	p.logLevel = logLevel
	return p
}

func (p *ServerConfig) SetNumOfThreads(numOfThreads int) *ServerConfig {
	p.numOfThreads = numOfThreads
	return p
}

func (p *ServerConfig) SetMaxNodeDepth(maxNodeDepth int16) *ServerConfig {
	p.maxNodeDepth = maxNodeDepth
	return p
}

func (p *ServerConfig) SetMaxCallDepth(maxCallDepth int16) *ServerConfig {
	p.maxCallDepth = maxCallDepth
	return p
}

func (p *ServerConfig) SetThreadBufferSize(
	threadBufferSize uint32,
) *ServerConfig {
	p.threadBufferSize = threadBufferSize
	return p
}

func (p *ServerConfig) SetCloseTimeout(
	closeTimeout time.Duration,
) *ServerConfig {
	p.closeTimeout = closeTimeout
	return p
}

func (p *ServerConfig) SetactionCache(
	actionCache rpc.ActionCache,
) *ServerConfig {
	p.actionCache = actionCache
	return p
}

func (p *ServerConfig) SetSession(session *SessionConfig) *ServerConfig {
	if session == nil {
		session = GetDefaultSessionConfig()
	}

	p.session = session
	return p
}

func (p *ServerConfig) clone() *ServerConfig {
	return &ServerConfig{
		logToScreen:      p.logToScreen,
		logFile:          p.logFile,
		logLevel:         p.logLevel,
		numOfThreads:     p.numOfThreads,
		maxNodeDepth:     p.maxNodeDepth,
		maxCallDepth:     p.maxCallDepth,
		threadBufferSize: p.threadBufferSize,
		closeTimeout:     p.closeTimeout,
		actionCache:      p.actionCache,
		session:          p.session.clone(),
	}
}
