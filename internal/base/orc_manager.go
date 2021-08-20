package base

import (
	"sync"
	"sync/atomic"
)

const (
	orcStatusNone    = uint32(0)
	orcStatusOpened  = uint32(1)
	orcStatusRunning = uint32(2)
	orcStatusDone    = uint32(3)
	orcStatusClosing = uint32(4)
	orcStatusClosed  = uint32(5)
)

// IORCService ...
type IORCService interface {
	Open() bool
	Run()
	Close() bool
}

// ORCManager ...
type ORCManager struct {
	status      uint32
	statusCond  sync.Cond
	runningCond sync.Cond
	mu          sync.Mutex
}

// NewORCManager ...
func NewORCManager() *ORCManager {
	return &ORCManager{}
}

func (p *ORCManager) getStatus() uint32 {
	return atomic.LoadUint32(&p.status)
}

func (p *ORCManager) setStatus(status uint32) {
	if p.getStatus() != status {
		atomic.StoreUint32(&p.status, status)
		p.statusCond.Broadcast()
	}
}

func (p *ORCManager) waitForStatusChange() {
	if p.statusCond.L == nil {
		p.statusCond.L = &p.mu
	}

	p.statusCond.Wait()
}

func (p *ORCManager) decRuningCount(ptrRunningCount *int64) {
	atomic.AddInt64(ptrRunningCount, -1)
	p.runningCond.Broadcast()
}

func (p *ORCManager) waitForRunningCountChange() {
	if p.runningCond.L == nil {
		p.runningCond.L = &p.mu
	}

	p.runningCond.Wait()
}

func (p *ORCManager) IsRunning() bool {
	return p.getStatus() == orcStatusRunning
}

func (p *ORCManager) IsClosing() bool {
	return p.getStatus() == orcStatusClosing
}

func (p *ORCManager) IsClosed() bool {
	return p.getStatus() == orcStatusClosed
}

// Open ...
func (p *ORCManager) Open(onOpen func() bool) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.getStatus() == orcStatusNone {
		if onOpen != nil && !onOpen() {
			p.setStatus(orcStatusClosed)
			return false
		}

		p.setStatus(orcStatusOpened)
		return true
	}

	return false
}

// Run ...
func (p *ORCManager) Run(onRuns ...func(isRunning func() bool)) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.getStatus() == orcStatusOpened {
		p.setStatus(orcStatusRunning)
		numOfRunning := int64(len(onRuns))

		for i := 0; i < len(onRuns); i++ {
			go func(onRun func(isRunning func() bool)) {
				defer func() {
					p.mu.Lock()
					defer p.mu.Unlock()

					p.decRuningCount(&numOfRunning)
				}()

				if onRun != nil {
					onRun(p.IsRunning)
				}
			}(onRuns[i])
		}

		// wait for all running routine done, then lock it again
		for atomic.LoadInt64(&numOfRunning) > 0 {
			p.waitForRunningCountChange()
		}

		switch p.getStatus() {
		case orcStatusRunning:
			p.setStatus(orcStatusDone)
		case orcStatusClosing:
			p.setStatus(orcStatusClosed)
		}
	}
}

// Close close the ORCManager
// willClose will be called immediately if ORCManager is opened correctly.
// didClose will be called when ORCManager is opened correctly and all onRun
// functions (if present) have completed
func (p *ORCManager) Close(willClose func(), didClose func()) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	status := p.getStatus()

	if status == orcStatusRunning {
		p.setStatus(orcStatusClosing)

		if willClose != nil {
			willClose()
		}

		for p.getStatus() != orcStatusClosed {
			p.waitForStatusChange()
		}

		if didClose != nil {
			didClose()
		}

		return true
	} else if status == orcStatusOpened || status == orcStatusDone {
		if willClose != nil {
			willClose()
		}

		p.setStatus(orcStatusClosed)

		if didClose != nil {
			didClose()
		}

		return true
	} else if status == orcStatusNone {
		p.setStatus(orcStatusClosed)
		return true
	} else {
		return false
	}
}
