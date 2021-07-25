package server

import (
	"runtime"
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

type testActionCache struct{}

func (p *testActionCache) Get(_ string) rpc.ActionCacheFunc {
	return nil
}

func TestGetDefaultSessionConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.numOfChannels).Equals(32)
		assert(v.transLimit).Equals(4 * 1024 * 1024)
		assert(v.heartbeatInterval).Equals(4 * time.Second)
		assert(v.heartbeatTimeout).Equals(8 * time.Second)
		assert(v.serverMaxSessions).Equals(10240000)
		assert(v.serverSessionTimeout).Equals(120 * time.Second)
		assert(v.serverReadBufferSize).Equals(1200)
		assert(v.serverWriteBufferSize).Equals(1200)
		assert(v.serverCacheTimeout).Equals(10 * time.Second)
	})
}

func TestSessionConfig_SetNumOfChannels(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetNumOfChannels(1)).Equals(v)
		assert(v.numOfChannels).Equals(1)
	})
}

func TestSessionConfig_SetTransLimit(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetTransLimit(1)).Equals(v)
		assert(v.transLimit).Equals(1)
	})
}

func TestSessionConfig_SetHeartbeatInterval(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetHeartbeatInterval(time.Second)).Equals(v)
		assert(v.heartbeatInterval).Equals(time.Second)
	})
}

func TestSessionConfig_SetHeartbeatTimeout(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetHeartbeatTimeout(time.Second)).Equals(v)
		assert(v.heartbeatTimeout).Equals(time.Second)
	})
}

func TestSessionConfig_SetServerMaxSessions(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetServerMaxSessions(1)).Equals(v)
		assert(v.serverMaxSessions).Equals(1)
	})
}

func TestSessionConfig_SetServerSessionTimeout(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetServerSessionTimeout(time.Second)).Equals(v)
		assert(v.serverSessionTimeout).Equals(time.Second)
	})
}

func TestSessionConfig_SetServerReadBufferSize(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetServerReadBufferSize(1)).Equals(v)
		assert(v.serverReadBufferSize).Equals(1)
	})
}

func TestSessionConfig_SetServerWriteBufferSize(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetServerWriteBufferSize(1)).Equals(v)
		assert(v.serverWriteBufferSize).Equals(1)
	})
}

func TestSessionConfig_SetServerCacheTimeout(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.SetServerCacheTimeout(time.Second)).Equals(v)
		assert(v.serverCacheTimeout).Equals(time.Second)
	})
}

func TestSessionConfig_clone(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultSessionConfig()
		assert(v.clone()).Equals(v)
	})
}

func TestGetDefaultServerConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.logToScreen).IsTrue()
		assert(v.logFile).Equals("")
		assert(v.logLevel).Equals(base.ErrorLogAll)
		assert(v.numOfThreads).Equals(base.MinInt(runtime.NumCPU(), 64) * 16384)
		assert(v.maxNodeDepth).Equals(int16(128))
		assert(v.maxCallDepth).Equals(int16(128))
		assert(v.threadBufferSize).Equals(uint32(2048))
		assert(v.closeTimeout).Equals(5 * time.Second)
		assert(v.actionCache).Equals(nil)
		assert(v.session).Equals(GetDefaultSessionConfig())
	})
}

func TestServerConfig_SetLogToScreen(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetLogToScreen(false)).Equals(v)
		assert(v.logToScreen).Equals(false)
	})
}

func TestServerConfig_SetLogFile(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetLogFile("file")).Equals(v)
		assert(v.logFile).Equals("file")
	})
}

func TestServerConfig_SetLogLevel(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetLogLevel(base.ErrorLevelError)).Equals(v)
		assert(v.logLevel).Equals(base.ErrorLevelError)
	})
}

func TestServerConfig_SetNumOfThreads(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetNumOfThreads(1)).Equals(v)
		assert(v.numOfThreads).Equals(1)
	})
}

func TestServerConfig_SetMaxNodeDepth(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetMaxNodeDepth(1)).Equals(v)
		assert(v.maxNodeDepth).Equals(int16(1))
	})
}

func TestServerConfig_SetMaxCallDepth(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetMaxCallDepth(1)).Equals(v)
		assert(v.maxCallDepth).Equals(int16(1))
	})
}

func TestServerConfig_SetThreadBufferSize(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetThreadBufferSize(1)).Equals(v)
		assert(v.threadBufferSize).Equals(uint32(1))
	})
}

func TestServerConfig_SetCloseTimeout(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetCloseTimeout(time.Second)).Equals(v)
		assert(v.closeTimeout).Equals(time.Second)
	})
}

func TestServerConfig_SetactionCache(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.SetactionCache(&testActionCache{})).Equals(v)
		assert(v.actionCache).IsNotNil()
	})
}

func TestServerConfig_SetSession(t *testing.T) {
	t.Run("session is nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		v.session.heartbeatInterval = 0
		assert(v.SetSession(nil)).Equals(v)
		assert(v.session).Equals(GetDefaultSessionConfig())
	})

	t.Run("session is not nil", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		session := GetDefaultSessionConfig().SetHeartbeatTimeout(time.Second)
		assert(v.SetSession(session)).Equals(v)
		assert(v.session).Equals(session)
	})
}

func TestServerConfig_clone(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := GetDefaultServerConfig()
		assert(v.clone()).Equals(v)
	})
}
