package server

import (
	"testing"
	"time"

	"github.com/rpccloud/rpc/internal/base"
)

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
	})
}
