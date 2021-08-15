package router

// import (
// 	"crypto/tls"
// 	"net"
// 	"strings"
// 	"sync"
// 	"time"

// 	"github.com/rpccloud/rpc/internal/base"
// 	"github.com/rpccloud/rpc/internal/rpc"
// )

// // Server ...
// type Server struct {
// 	addr           string
// 	tlsConfig      *tls.Config
// 	orcManager     *base.ORCManager
// 	streamReceiver rpc.IStreamReceiver
// 	id             *base.GlobalID
// 	router         *Router
// 	ln             net.Listener
// 	sync.Mutex
// }

// // NewServer ...
// func NewServer(
// 	addr string,
// 	tlsConfig *tls.Config,
// 	streamReceiver rpc.IStreamReceiver,
// ) *Server {
// 	ret := &Server{
// 		addr:           addr,
// 		tlsConfig:      tlsConfig,
// 		orcManager:     base.NewORCManager(),
// 		streamReceiver: streamReceiver,
// 		id:             nil,
// 		router:         NewRouter(streamReceiver),
// 		ln:             nil,
// 	}

// 	return ret
// }

// // Open ...
// func (p *Server) Open() bool {
// 	return p.orcManager.Open(func() bool {
// 		e := error(nil)
// 		if p.tlsConfig == nil {
// 			p.ln, e = net.Listen("tcp", p.addr)
// 		} else {
// 			p.ln, e = tls.Listen("tcp", p.addr, p.tlsConfig)
// 		}
// 		if e != nil {
// 			p.streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(
// 				base.ErrRouterConnListen.AddDebug(e.Error()),
// 			))
// 			return false
// 		}

// 		return true
// 	})
// }

// // Run ...
// func (p *Server) Run() bool {
// 	return p.orcManager.Run(func(isRunning func() bool) bool {

// 		for isRunning() {
// 			conn, e := p.ln.Accept()
// 			if e != nil {
// 				isCloseErr := !isRunning() &&
// 					strings.HasSuffix(e.Error(), base.ErrNetClosingSuffix)

// 				if !isCloseErr {
// 					p.streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(
// 						base.ErrRouterConnConnect.AddDebug(e.Error()),
// 					))

// 					base.WaitWhileRunning(
// 						base.TimeNow().UnixNano(),
// 						isRunning,
// 						500*time.Millisecond,
// 					)
// 				}
// 			} else {
// 				go p.onConnect(conn)
// 			}
// 		}

// 		return true
// 	})
// }

// func (p *Server) onConnect(conn net.Conn) {
// 	if err := p.router.AddConn(conn); err != nil {
// 		p.streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(err))
// 	}
// }

// // Close ...
// func (p *Server) Close() bool {
// 	return p.orcManager.Close(func() bool {
// 		if e := p.ln.Close(); e != nil {
// 			p.streamReceiver.OnReceiveStream(rpc.MakeSystemErrorStream(
// 				base.ErrRouterConnClose.AddDebug(e.Error()),
// 			))
// 		}

// 		return true
// 	}, func() {
// 		p.ln = nil
// 	})
// }
