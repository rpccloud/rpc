package router

import (
	"crypto/tls"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rpccloud/rpc/internal/base"
	"github.com/rpccloud/rpc/internal/rpc"
)

type RouterKind int

const (
	RouterKindR  RouterKind = 1
	RouterKindM  RouterKind = 2
	RouterKindRM RouterKind = 3
)

// Router ...
type Router struct {
	kind        RouterKind
	ln          net.Listener
	clusterAddr string
	upLink      string
	streamHub   *rpc.StreamHub
	orcManager  *base.ORCManager
	sync.Mutex
}

// NewRouter ...
func NewRouter(
	listenAddr string,
	listenTlsConfig *tls.Config,
	logToScreen bool,
	logFile string,
	logLevel base.ErrorLevel,
) *Router {
	ret := &Router{
		kind:        RouterKindRM,
		ln:          nil,
		clusterAddr: "",
		upLink:      "",
		streamHub:   nil,
		orcManager:  base.NewORCManager(),
	}

	ret.streamHub = rpc.NewStreamHub(
		logToScreen,
		logFile,
		logLevel,
		rpc.StreamHubCallback{
			OnRPCRequestStream: func(stream *rpc.Stream) {
				ret.route(stream)
			},
			OnRPCResponseOKStream: func(stream *rpc.Stream) {
				ret.route(stream)
			},
			OnRPCResponseErrorStream: func(stream *rpc.Stream) {
				ret.route(stream)
			},
			OnRPCBoardCastStream: func(stream *rpc.Stream) {
				ret.route(stream)
			},
			OnSystemErrorReportStream: func(
				sessionID uint64,
				err *base.Error,
			) {
				// ignore
			},
		},
	)

	ret.orcManager.Open(func() bool {
		e := error(nil)

		if listenTlsConfig == nil {
			ret.ln, e = net.Listen("tcp", listenAddr)
		} else {
			ret.ln, e = tls.Listen("tcp", listenAddr, listenTlsConfig)
		}

		if e != nil {
			ret.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
				base.ErrRouterConnListen.AddDebug(e.Error()),
			))
			return false
		}

		return true
	})

	go func() {
		ret.orcManager.Run(func(isRunning func() bool) {
			waitTimerCloseCH := make(chan bool)
			connProcessing := int64(0)

			// onTimer go routine
			go func() {
				sequence := uint64(0)
				for isRunning() {
					sequence++
					ret.onTimer(sequence)
					time.Sleep(100 * time.Millisecond)
				}
				waitTimerCloseCH <- true
			}()

			// accept connections
			for isRunning() {
				conn, e := ret.ln.Accept()

				if e != nil {
					isCloseErr := ret.orcManager.IsClosing() &&
						strings.HasSuffix(e.Error(), base.ErrNetClosingSuffix)

					if !isCloseErr {
						ret.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
							base.ErrRouterConnConnect.AddDebug(e.Error()),
						))

						base.WaitWhileRunning(
							base.TimeNow().UnixNano(),
							isRunning,
							500*time.Millisecond,
						)
					}
				} else {
					atomic.AddInt64(&connProcessing, 1)

					go func() {
						defer func() {
							atomic.AddInt64(&connProcessing, -1)
						}()

						if err := ret.addConn(conn); err != nil {
							ret.streamHub.OnReceiveStream(
								rpc.MakeSystemErrorStream(err),
							)
						}
					}()
				}
			}

			// wait until all addConn done
			for atomic.LoadInt64(&connProcessing) != 0 {
				time.Sleep(20 * time.Millisecond)
			}

			// wait onTimer go routine done
			<-waitTimerCloseCH
		})
	}()

	return ret
}

func (p *Router) route(stream *rpc.Stream) {

}

func (p *Router) onTimer(sequence uint64) {

}

func (p *Router) SetUpLink(upLink string) {

}

func (p *Router) SetClusterAddr(clusterAddr string) {

}

func (p *Router) Close() bool {
	return p.orcManager.Close(
		func() {
			if e := p.ln.Close(); e != nil {
				p.streamHub.OnReceiveStream(rpc.MakeSystemErrorStream(
					base.ErrRouterConnListen.AddDebug(e.Error()),
				))
			}
		},
		func() {
			p.ln = nil
			p.streamHub.Close()
		},
	)
}

// // OnReceiveStream ...
// func (p *Router) OnReceiveStream(s *rpc.Stream) {
// 	p.streamReceiver.OnReceiveStream(s)
// }

func (p *Router) addConn(conn net.Conn) *base.Error {
	return nil
	// var buffer [32]byte
	// n, err := connReadBytes(conn, time.Second, buffer[:])

	// if err != nil || n != 32 {
	// 	_ = conn.Close()
	// 	return err
	// }

	// if binary.LittleEndian.Uint16(buffer[2:]) != channelActionInit {
	// 	_ = conn.Close()
	// 	return base.ErrRouterConnProtocol
	// }

	// slotID := binary.LittleEndian.Uint64(buffer[6:])

	// p.Lock()
	// slot, ok := p.slotMap[slotID]
	// if !ok {
	// 	slot = NewSlot(nil, p)
	// 	p.slotMap[slotID] = slot
	// }
	// p.Unlock()

	// return slot.addSlaveConn(conn, buffer)
}

// func (p *Router) delSlot(id uint64) {
// 	p.Lock()
// 	defer p.Unlock()

// 	if slot, ok := p.slotMap[id]; ok {
// 		delete(p.slotMap, id)
// 		slot.Close()
// 	}
// }
