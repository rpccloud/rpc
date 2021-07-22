package rpc

import (
	"testing"

	"github.com/rpccloud/rpc/internal/base"
)

func TestNewServiceMeta(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := NewService()
		assert(NewServiceMeta("test", service, "debug", nil)).
			Equals(&ServiceMeta{
				name:     "test",
				service:  service,
				fileLine: "debug",
			})
	})
}

func TestNewService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		service := NewService()
		assert(service).IsNotNil()
		assert(len(service.children)).Equals(0)
		assert(len(service.actions)).Equals(0)
	})
}

func TestService_AddChildService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		child := NewService()
		service, fileLine := NewService().
			AddChildService("ch", child, nil), base.GetFileLine(0)
		assert(service).IsNotNil()
		assert(len(service.children)).Equals(1)
		assert(len(service.actions)).Equals(0)
		assert(service.children[0].name).Equals("ch")
		assert(service.children[0].service).Equals(child)
		assert(service.children[0].fileLine).Equals(fileLine)
	})
}

func TestService_On(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		getFL := base.GetFileLine
		service, fileLine := NewService().On("sayHello", 2345), getFL(0)
		assert(service).IsNotNil()
		assert(len(service.children)).Equals(0)
		assert(len(service.actions)).Equals(1)
		assert(service.actions[0].name).Equals("sayHello")
		assert(service.actions[0].handler).Equals(2345)
		assert(service.actions[0].fileLine).Equals(fileLine)
	})
}
