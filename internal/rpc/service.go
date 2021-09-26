package rpc

import (
	"sync"

	"github.com/rpccloud/rpc/internal/base"
)

// ActionMeta ...
type ActionMeta struct {
	name     string      // action name
	handler  interface{} // action handler
	fileLine string      // where the action add in source file
}

// ServiceMeta ...
type ServiceMeta struct {
	name     string   // the name of child service
	service  *Service // the real service
	fileLine string   // where the service add in source file
	data     Map
}

// NewServiceMeta ...
func NewServiceMeta(
	name string,
	service *Service,
	fileLine string,
	data Map,
) *ServiceMeta {
	return &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: fileLine,
		data:     data,
	}
}

// Service ...
type Service struct {
	children []*ServiceMeta // all the children node meta pointer
	actions  []*ActionMeta  // all the actions meta pointer
	mu       sync.Mutex
}

// NewService define a new service
func NewService() *Service {
	return &Service{
		children: nil,
		actions:  nil,
	}
}

// AddChildService ...
func (p *Service) AddChildService(
	name string,
	service *Service,
	data Map,
) *Service {
	p.mu.Lock()
	defer p.mu.Unlock()

	// add child meta
	p.children = append(p.children, &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: base.GetFileLine(1),
		data:     data,
	})
	return p
}

// On ...
func (p *Service) On(
	name string,
	handler interface{},
) *Service {
	p.mu.Lock()
	defer p.mu.Unlock()

	// add action meta
	p.actions = append(p.actions, &ActionMeta{
		name:     name,
		handler:  handler,
		fileLine: base.GetFileLine(1),
	})
	return p
}
