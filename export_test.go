package rpc

import (
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func TestNewService(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewService(nil)).IsNotNil()
	})
}

func TestNewServer(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewServer(nil)).IsNotNil()
	})
}

func TestNewClient(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		v := NewClient("ws", "127.0.0.1", "", nil, 1500, 1500, nil)
		defer v.Close()
		assert(v).IsNotNil()
	})
}

func TestGetTLSServerConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetServerTLSConfig("errorFile", "errorFile")).
			Equals(base.GetServerTLSConfig("errorFile", "errorFile"))
	})
}

func TestGetTLSClientConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetClientTLSConfig(true, nil)).
			Equals(base.GetClientTLSConfig(true, nil))
	})
}
