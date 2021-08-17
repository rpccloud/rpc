package adapter

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"path"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/rpccloud/rpc/internal/base"
)

func getFieldPointer(ptr interface{}, fieldName string) unsafe.Pointer {
	val := reflect.Indirect(reflect.ValueOf(ptr))
	return unsafe.Pointer(val.FieldByName(fieldName).UnsafeAddr())
}

func syncServerTestOpen(
	network string,
	isTLS bool,
	fakeAddrError bool,
	fakeConnectError bool,
) (*testSingleReceiver, bool) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	addr := ""
	if fakeAddrError {
		addr = "error-addr"
	} else {
		addr = "127.0.0.1:65432"
	}

	tlsClientConfig := (*tls.Config)(nil)
	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
		)
		tlsClientConfig, _ = base.GetTLSClientConfig(true, []string{
			path.Join(curDir, "_cert_", "ca", "ca.pem"),
		})
	}

	receiver := newTestSingleReceiver()

	v := NewSyncServerService(NewServerAdapter(
		false, network, addr, tlsServerConfig, 1200, 1200, receiver,
	))
	openOK := v.Open()
	go func() {
		v.Run()
	}()

	if (network == "ws" || network == "wss") && fakeConnectError {
		go func() {
			time.Sleep(100 * time.Millisecond)
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: tlsClientConfig,
				},
			}
			if isTLS {
				_, _ = client.Get("https://127.0.0.1:65432")
			} else {
				_, _ = client.Get("http://127.0.0.1:65432")
			}
		}()
	}

	if fakeAddrError || fakeConnectError {
		for receiver.GetOnErrorCount() == 0 {
			time.Sleep(10 * time.Millisecond)
		}
	}

	v.Close()
	return receiver, openOK
}

type fakeTestRunListener struct {
	emulateAcceptError bool
	emulateCLoseError  bool
	ln                 net.Listener
}

func (p *fakeTestRunListener) Accept() (net.Conn, error) {
	if p.emulateAcceptError {
		return nil, io.EOF
	}

	return p.ln.Accept()
}

func (p *fakeTestRunListener) Close() error {
	if p.emulateCLoseError {
		_ = p.ln.Close()
		return io.EOF
	}

	return p.ln.Close()
}

func (p *fakeTestRunListener) Addr() net.Addr {
	return p.ln.Addr()
}

func syncServerTestRun(
	network string,
	isTLS bool,
	fakeError bool,
) *testSingleReceiver {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	tlsClientConfig := (*tls.Config)(nil)
	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
		)
		tlsClientConfig, _ = base.GetTLSClientConfig(true, []string{
			path.Join(curDir, "_cert_", "ca", "ca.pem"),
		})
	}

	waitCH := make(chan bool)
	receiver := newTestSingleReceiver()
	server := NewSyncServerService(NewServerAdapter(
		false, network, "127.0.0.1:65432", tlsServerConfig,
		1200, 1200, receiver,
	))
	server.Open()
	fakeLN := &fakeTestRunListener{
		emulateAcceptError: fakeError,
		emulateCLoseError:  false,
	}
	if strings.HasPrefix(network, "tcp") {
		fakeLN.ln = server.(*syncTCPServerService).ln
		server.(*syncTCPServerService).ln = fakeLN
	} else {
		fakeLN.ln = server.(*syncWSServerService).ln
		server.(*syncWSServerService).ln = fakeLN
	}

	go func() {
		go func() {
			if fakeError {
				for receiver.GetOnErrorCount() == 0 {
					time.Sleep(10 * time.Millisecond)
				}
			} else {
				for receiver.GetOnOpenCount() == 0 {
					time.Sleep(10 * time.Millisecond)
				}
			}

			waitCH <- true
		}()

		server.Run()
		waitCH <- true
	}()

	// Start client
	go func() {
		client := NewSyncClientService(NewClientAdapter(
			network, "127.0.0.1:65432", tlsClientConfig,
			1200, 1200, newTestSingleReceiver(),
		))
		client.Open()
		go func() {
			client.Run()
		}()

		// wait server signal ...
		if fakeError {
			for receiver.GetOnErrorCount() == 0 {
				time.Sleep(10 * time.Millisecond)
			}
		} else {
			for receiver.GetOnOpenCount() == 0 {
				time.Sleep(10 * time.Millisecond)
			}
		}
		// close
		client.Close()
	}()

	<-waitCH
	server.Close()

	for receiver.GetOnOpenCount() != receiver.GetOnCloseCount() {
		time.Sleep(50 * time.Millisecond)
	}

	<-waitCH
	return receiver
}

func syncServerTestClose(
	network string,
	isTLS bool,
	fakeError bool,
) (*testSingleReceiver, bool) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
		)
	}

	receiver := newTestSingleReceiver()
	v := NewSyncServerService(NewServerAdapter(
		false, network, "127.0.0.1:65432", tlsServerConfig,
		1200, 1200, receiver,
	))
	v.Open()

	fakeLN := &fakeTestRunListener{
		emulateAcceptError: false,
		emulateCLoseError:  fakeError,
	}
	if strings.HasPrefix(network, "tcp") {
		fakeLN.ln = v.(*syncTCPServerService).ln
		v.(*syncTCPServerService).ln = fakeLN
	} else {
		fakeLN.ln = v.(*syncWSServerService).ln
		v.(*syncWSServerService).ln = fakeLN
	}

	go func() {
		v.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	if fakeError {
		_ = fakeLN.ln.Close()
		if network == "ws" || network == "wss" {
			httpServer := v.(*syncWSServerService).server
			ptrMU := (*sync.Mutex)(getFieldPointer(httpServer, "mu"))
			ptrListeners := (*map[*net.Listener]struct{})(getFieldPointer(
				httpServer, "listeners",
			))
			ptrMU.Lock()
			fakeListener := net.Listener(&net.TCPListener{})
			(*ptrListeners)[&fakeListener] = struct{}{}
			ptrMU.Unlock()
		}
	}

	closeOK := v.Close()
	return receiver, closeOK
}

func syncClientTest(
	isTLS bool,
	serverNetwork string,
	clientNetwork string,
	addr string,
) (*testSingleReceiver, *syncClientService) {
	_, curFile, _, _ := runtime.Caller(0)
	curDir := path.Dir(curFile)

	tlsClientConfig := (*tls.Config)(nil)
	tlsServerConfig := (*tls.Config)(nil)

	if isTLS {
		tlsServerConfig, _ = base.GetTLSServerConfig(
			path.Join(curDir, "_cert_", "server", "server.pem"),
			path.Join(curDir, "_cert_", "server", "server-key.pem"),
		)
		tlsClientConfig, _ = base.GetTLSClientConfig(true, []string{
			path.Join(curDir, "_cert_", "ca", "ca.pem"),
		})
	}

	server := NewSyncServerService(NewServerAdapter(
		false, serverNetwork, addr, tlsServerConfig,
		1200, 1200, newTestSingleReceiver(),
	))
	server.Open()
	defer server.Close()
	go func() {
		server.Run()
	}()

	time.Sleep(100 * time.Millisecond)

	clientReceiver := newTestSingleReceiver()
	client := &syncClientService{
		adapter: NewClientAdapter(
			clientNetwork, addr, tlsClientConfig,
			1200, 1200, clientReceiver,
		),
		orcManager: base.NewORCManager(),
	}

	client.Open()
	defer client.Close()

	go func() {
		client.Run()
	}()

	for clientReceiver.GetOnOpenCount() == 0 &&
		clientReceiver.GetOnErrorCount() == 0 {
		time.Sleep(10 * time.Millisecond)
	}

	return clientReceiver, client
}

func TestNewSyncClientService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for _, network := range []string{
			"tcp", "tcp4", "tcp6", "ws", "wss",
		} {
			adapter := NewClientAdapter(
				network, "localhost", nil, 1200, 1200, newTestSingleReceiver(),
			)

			service := NewSyncClientService(adapter).(*syncClientService)

			assert(service.adapter).Equals(adapter)
			assert(service.conn).IsNil()
			assert(service.orcManager).IsNotNil()
		}
	})

	t.Run("protocol error", func(t *testing.T) {
		assert := base.NewAssert(t)

		receiver := newTestSingleReceiver()
		adapter := NewClientAdapter(
			"err", "localhost", nil, 1200, 1200, receiver,
		)

		service := NewSyncClientService(adapter)
		assert(service).IsNil()
		assert(receiver.GetOnErrorCount()).Equals(1)
		assert(receiver.GetError()).
			Equals(base.ErrUnsupportedProtocol.AddDebug(
				"unsupported protocol err",
			))
	})
}

func TestNewSyncServerService(t *testing.T) {
	t.Run("test ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		for _, network := range []string{"tcp", "tcp4", "tcp6"} {
			adapter := NewClientAdapter(
				network, "localhost", nil, 1200, 1200, newTestSingleReceiver(),
			)

			service := NewSyncServerService(adapter).(*syncTCPServerService)
			assert(service.adapter).Equals(adapter)
			assert(service.ln).IsNil()
			assert(service.orcManager).IsNotNil()
		}

		for _, network := range []string{"ws", "wss"} {
			adapter := NewClientAdapter(
				network, "localhost", nil, 1200, 1200, newTestSingleReceiver(),
			)

			service := NewSyncServerService(adapter).(*syncWSServerService)
			assert(service.adapter).Equals(adapter)
			assert(service.ln).IsNil()
			assert(service.server).IsNil()
			assert(service.orcManager).IsNotNil()
		}
	})

	t.Run("protocol error", func(t *testing.T) {
		assert := base.NewAssert(t)

		receiver := newTestSingleReceiver()
		adapter := NewClientAdapter(
			"err", "localhost", nil, 1200, 1200, receiver,
		)

		service := NewSyncServerService(adapter)
		assert(service).IsNil()
		assert(receiver.GetOnErrorCount()).Equals(1)
		assert(receiver.GetError()).
			Equals(base.ErrUnsupportedProtocol.AddDebug(
				"unsupported protocol err",
			))
	})
}

func TestSyncTCPServerService_Open(t *testing.T) {
	t.Run("tcp open error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", false, true, false)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("tcp open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", false, false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("tls open error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", true, true, false)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("tls open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("tcp", false, false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncTCPServerService_Run(t *testing.T) {
	t.Run("tcp run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("tcp", false, true)
		assert(receiver.GetError()).
			Equals(base.ErrSyncTCPServerServiceAccept.AddDebug(io.EOF.Error()))
		assert(receiver.GetOnOpenCount()).Equals(0)
		assert(receiver.GetOnCloseCount()).Equals(0)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount() > 0).IsTrue()
	})

	t.Run("tcp run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("tcp", false, false)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equals(1)
		assert(receiver.GetOnCloseCount()).Equals(1)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount()).Equals(0)
	})

	t.Run("tls run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("tcp", true, true)
		assert(receiver.GetError()).
			Equals(base.ErrSyncTCPServerServiceAccept.AddDebug(io.EOF.Error()))
		assert(receiver.GetOnOpenCount()).Equals(0)
		assert(receiver.GetOnCloseCount()).Equals(0)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount() > 0).IsTrue()
	})

	t.Run("tls run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("tcp", true, false)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equals(1)
		assert(receiver.GetOnCloseCount()).Equals(1)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount()).Equals(0)
	})
}

func TestSyncTCPServerService_Close(t *testing.T) {
	t.Run("tcp close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("tcp", false, true)
		assert(ok).IsTrue()
		assert(receiver.GetOnErrorCount() >= 1).IsTrue()
	})

	t.Run("tcp close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("tcp", false, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("tls close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("tcp", true, true)
		assert(ok).IsTrue()
		assert(receiver.GetOnErrorCount() >= 1).IsTrue()
	})

	t.Run("tls close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("tcp", true, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncWSServerService_Open(t *testing.T) {
	t.Run("ws open addr error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("ws", false, true, false)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("ws open connect error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("ws", false, false, true)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("ws open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("ws", false, false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("wss open addr error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("wss", true, true, false)
		assert := base.NewAssert(t)
		assert(ok).IsFalse()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("wss open connect error", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("wss", true, false, true)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNotNil()
	})

	t.Run("wss open ok", func(t *testing.T) {
		receiver, ok := syncServerTestOpen("wss", true, false, false)
		assert := base.NewAssert(t)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncWSServerService_Run(t *testing.T) {
	t.Run("ws run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("ws", false, true)
		assert(receiver.GetError()).
			Equals(base.ErrSyncWSServerServiceServe.AddDebug(io.EOF.Error()))
		assert(receiver.GetOnOpenCount()).Equals(0)
		assert(receiver.GetOnCloseCount()).Equals(0)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount() > 0).IsTrue()
	})

	t.Run("ws run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("ws", false, false)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equals(1)
		assert(receiver.GetOnCloseCount()).Equals(1)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount()).Equals(0)
	})

	t.Run("tls run error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("tcp", true, true)
		assert(receiver.GetError()).
			Equals(base.ErrSyncTCPServerServiceAccept.AddDebug(io.EOF.Error()))
		assert(receiver.GetOnOpenCount()).Equals(0)
		assert(receiver.GetOnCloseCount()).Equals(0)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount() > 0).IsTrue()
	})

	t.Run("tls run ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver := syncServerTestRun("tcp", true, false)
		assert(receiver.GetError()).IsNil()
		assert(receiver.GetOnOpenCount()).Equals(1)
		assert(receiver.GetOnCloseCount()).Equals(1)
		assert(receiver.GetOnStreamCount()).Equals(0)
		assert(receiver.GetOnErrorCount()).Equals(0)
	})

}

func TestSyncWSServerService_Close(t *testing.T) {
	t.Run("ws close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("ws", false, true)
		assert(ok).IsTrue()
		assert(receiver.GetOnErrorCount() >= 1).IsTrue()
	})

	t.Run("ws close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("ws", false, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})

	t.Run("wss close error", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("wss", true, true)
		assert(ok).IsTrue()
		assert(receiver.GetOnErrorCount() >= 1).IsTrue()
	})

	t.Run("wss close ok", func(t *testing.T) {
		assert := base.NewAssert(t)
		receiver, ok := syncServerTestClose("wss", true, false)
		assert(ok).IsTrue()
		assert(receiver.GetError()).IsNil()
	})
}

func TestSyncClientService(t *testing.T) {
	type testItem struct {
		network string
		isTLS   bool
	}
	addr := "localhost:65437"

	t.Run("tcp4 ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		for _, it := range []testItem{
			{network: "tcp", isTLS: false},
			{network: "tcp", isTLS: true},
			{network: "tcp4", isTLS: false},
			{network: "tcp4", isTLS: true},
			{network: "tcp6", isTLS: false},
			{network: "tcp6", isTLS: true},
			{network: "ws", isTLS: false},
			{network: "wss", isTLS: true},
		} {
			receiver, client := syncClientTest(
				it.isTLS, it.network, it.network, addr,
			)
			assert(receiver.GetOnOpenCount()).Equals(1)
			assert(receiver.GetOnCloseCount()).Equals(1)
			assert(receiver.GetOnErrorCount()).Equals(0)
			assert(receiver.GetOnStreamCount()).Equals(0)
			assert(client.conn).IsNil()
		}
	})

	t.Run("tcp4 addr error", func(t *testing.T) {
		assert := base.NewAssert(t)

		for _, it := range []testItem{
			{network: "tcp", isTLS: false},
			{network: "tcp", isTLS: true},
			{network: "tcp4", isTLS: false},
			{network: "tcp4", isTLS: true},
			{network: "tcp6", isTLS: false},
			{network: "tcp6", isTLS: true},
			{network: "ws", isTLS: false},
			{network: "wss", isTLS: true},
		} {
			receiver, client := syncClientTest(
				it.isTLS, it.network, it.network, "addr-error",
			)

			assert(strings.Contains(
				receiver.GetError().GetMessage(),
				"dial",
			)).IsTrue()

			assert(receiver.GetOnOpenCount()).Equals(0)
			assert(receiver.GetOnCloseCount()).Equals(0)
			assert(receiver.GetOnErrorCount()).Equals(1)
			assert(receiver.GetOnStreamCount()).Equals(0)
			assert(client.conn).IsNil()
		}
	})

	t.Run("unsupported protocol", func(t *testing.T) {
		assert := base.NewAssert(t)

		for _, it := range []testItem{
			{network: "err", isTLS: false},
			{network: "err", isTLS: true},
		} {
			receiver, client := syncClientTest(
				it.isTLS, "tcp", it.network, addr,
			)

			assert(receiver.GetError()).
				Equals(base.ErrUnsupportedProtocol.AddDebug(
					"unsupported protocol err",
				))
			assert(receiver.GetOnOpenCount()).Equals(0)
			assert(receiver.GetOnCloseCount()).Equals(0)
			assert(receiver.GetOnErrorCount()).Equals(1)
			assert(receiver.GetOnStreamCount()).Equals(0)
			assert(client.conn).IsNil()
		}
	})
}
