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
	config   Map
}

// NewServiceMeta ...
func NewServiceMeta(
	name string,
	service *Service,
	fileLine string,
	config Map,
) *ServiceMeta {
	return &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: fileLine,
		config:   config,
	}
}

// Service ...
type Service struct {
	children []*ServiceMeta // all the children node meta pointer
	actions  []*ActionMeta  // all the actions meta pointer
	config   Map
	mu       sync.Mutex
}

// NewService define a new service
func NewService(config Map) *Service {
	return &Service{
		children: nil,
		actions:  nil,
		config:   config,
	}
}

// AddChildService ...
func (p *Service) AddChildService(
	name string,
	service *Service,
	config Map,
) *Service {
	p.mu.Lock()
	defer p.mu.Unlock()

	// add child meta
	p.children = append(p.children, &ServiceMeta{
		name:     name,
		service:  service,
		fileLine: base.GetFileLine(1),
		config:   config,
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
