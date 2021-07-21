package server

import (
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
)

func TestGetDefaultSessionConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		cfg := GetDefaultSessionConfig()
		assert(cfg.numOfChannels).Equal(32)
		assert(cfg.transLimit).Equal(4 * 1024 * 1024)
		assert(cfg.heartbeatInterval).Equal(4 * time.Second)
		assert(cfg.heartbeatTimeout).Equal(8 * time.Second)
		assert(cfg.serverMaxSessions).Equal(10240000)
		assert(cfg.serverSessionTimeout).Equal(120 * time.Second)
		assert(cfg.serverReadBufferSize).Equal(1200)
		assert(cfg.serverWriteBufferSize).Equal(1200)
		assert(cfg.serverCacheTimeout).Equal(10 * time.Second)
	})
}
