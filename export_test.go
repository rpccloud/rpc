package rpc

import (
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func TestNewService(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(NewService()).IsNotNil()
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
		v := NewClient("ws", "127.0.0.1", nil, 1500, 1500, nil)
		defer v.Close()
		assert(v).IsNotNil()
	})
}

func TestGetTLSServerConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetTLSServerConfig("errorFile", "errorFile")).
			Equal(base.GetTLSServerConfig("errorFile", "errorFile"))
	})
}

func TestGetTLSClientConfig(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		assert(GetTLSClientConfig(true, nil)).
			Equal(base.GetTLSClientConfig(true, nil))
	})
}
