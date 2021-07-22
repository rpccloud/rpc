package adapter

import (
	"context"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/rpccloud/rpc/internal/base"
)

func getTestConnections() (*syncWSConn, *syncWSConn, *http.Server) {
	clientConnCH := make(chan *syncWSConn, 1)
	serverConnCH := make(chan *syncWSConn, 1)
	httpServerCH := make(chan *http.Server, 1)

	go func() {
		server := &http.Server{
			Addr: ":12345",
			Handler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					conn, _, _, _ := ws.UpgradeHTTP(r, w)
					serverConnCH <- newSyncWSServerConn(conn)
				},
			),
		}

		httpServerCH <- server
		_ = server.ListenAndServe()
	}()

	go func() {
		time.Sleep(500 * time.Millisecond)
		conn, _, _, e := ws.DefaultDialer.Dial(
			context.Background(),
			"ws://127.0.0.1:12345",
		)
		if e != nil {
			panic(e)
		}
		clientConnCH <- newSyncWSClientConn(conn)
	}()

	return <-clientConnCH, <-serverConnCH, <-httpServerCH
}

func testConnReadAndWrite(
	totalBytes int,
	sliceSize int,
	sendConn net.Conn,
	receiveConn net.Conn,
) bool {
	returnRead := make(chan bool)
	returnWrite := make(chan bool)

	dataSend := make([]byte, totalBytes)
	for i := 0; i < totalBytes; i++ {
		dataSend[i] = byte(i)
	}

	go func() {
		_, e := sendConn.Write(dataSend)
		if e != nil {
			panic(e)
		}
	}()

	go func() {
		dataReceive := make([]byte, totalBytes)
		slice := make([]byte, sliceSize)
		pos := 0
		for pos < totalBytes {
			if n, e := receiveConn.Read(slice); e == nil {
				pos += copy(dataReceive[pos:], slice[:n])
			} else {
				panic(e)
			}
		}

		if reflect.DeepEqual(dataReceive, dataSend) {
			returnRead <- receiveConn.Close() == nil
			returnWrite <- sendConn.Close() == nil
		} else {
			_ = receiveConn.Close()
			_ = sendConn.Close()
			returnRead <- false
			returnWrite <- false
		}
	}()

	return <-returnRead && <-returnWrite
}

func TestSyncWSConn_Read(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		assert := base.NewAssert(t)
		clientConn, serverConn, server := getTestConnections()
		_ = serverConn.Close()
		buffer := make([]byte, 1024)
		assert(clientConn.Read(buffer)).Equals(-1, io.EOF)
		_ = server.Close()
	})

	t.Run("read ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		testCollection := [][2]int{
			{1, 1},
			{10, 1},
			{10, 3},
			{4 * 1024 * 1024, 1},
			{4 * 1024 * 1024, 1500},
		}

		for _, testSuite := range testCollection {
			clientConn, serverConn, server := getTestConnections()
			assert(testConnReadAndWrite(
				testSuite[0],
				testSuite[1],
				serverConn,
				clientConn,
			)).Equals(true)
			_ = server.Close()
		}
	})
}

func TestSyncWSConn_Write(t *testing.T) {
	t.Run("write error", func(t *testing.T) {
		assert := base.NewAssert(t)
		clientConn, _, server := getTestConnections()
		_ = clientConn.Close()
		buffer := make([]byte, 1024)
		n, e := clientConn.Write(buffer)
		assert(n).Equals(-1)
		assert(e).IsNotNil()
		assert(strings.HasSuffix(
			e.Error(),
			"use of closed network connection",
		)).Equals(true)
		_ = server.Close()
	})

	t.Run("write ok", func(t *testing.T) {
		assert := base.NewAssert(t)

		testCollection := [][2]int{
			{1, 1},
			{10, 1},
			{10, 3},
			{4 * 1024 * 1024, 1},
			{4 * 1024 * 1024, 1500},
		}

		for _, testSuite := range testCollection {
			clientConn, serverConn, server := getTestConnections()
			assert(testConnReadAndWrite(
				testSuite[0],
				testSuite[1],
				clientConn,
				serverConn,
			)).Equals(true)
			_ = server.Close()
		}
	})
}

func TestSyncWSConn_Close(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		assert := base.NewAssert(t)
		clientConn, serverConn, server := getTestConnections()
		assert(clientConn.Close()).IsNil()
		assert(serverConn.Close()).IsNil()
		_ = server.Close()
	})
}

func TestSyncWSConn_LocalAddr(t *testing.T) {
	assert := base.NewAssert(t)
	clientConn, serverConn, server := getTestConnections()
	assert(strings.HasPrefix(
		clientConn.LocalAddr().String(),
		"127.0.0.1:",
	)).Equals(true)
	assert(serverConn.LocalAddr().String()).Equals("127.0.0.1:12345")
	assert(serverConn.RemoteAddr().String()).
		Equals(clientConn.LocalAddr().String())
	_ = server.Close()
}

func TestSyncWSConn_RemoteAddr(t *testing.T) {
	assert := base.NewAssert(t)
	clientConn, serverConn, server := getTestConnections()

	assert(clientConn.RemoteAddr().String()).Equals("127.0.0.1:12345")
	assert(strings.HasPrefix(
		serverConn.RemoteAddr().String(),
		"127.0.0.1:",
	)).Equals(true)

	assert(serverConn.RemoteAddr().String()).
		Equals(clientConn.LocalAddr().String())
	_ = server.Close()
}

func TestSyncWSConn_SetDeadline(t *testing.T) {
	assert := base.NewAssert(t)
	clientConn, serverConn, server := getTestConnections()

	assert(
		clientConn.SetDeadline(base.TimeNow().Add(time.Second)),
	).IsNil()

	assert(
		serverConn.SetDeadline(base.TimeNow().Add(time.Second)),
	).IsNil()

	_ = server.Close()
}

func TestSyncWSConn_SetReadDeadline(t *testing.T) {
	assert := base.NewAssert(t)
	clientConn, serverConn, server := getTestConnections()

	assert(
		clientConn.SetReadDeadline(base.TimeNow().Add(time.Second)),
	).IsNil()

	assert(
		serverConn.SetReadDeadline(base.TimeNow().Add(time.Second)),
	).IsNil()

	_ = server.Close()
}

func TestSyncWSConn_SetWriteDeadline(t *testing.T) {
	assert := base.NewAssert(t)
	clientConn, serverConn, server := getTestConnections()

	assert(
		clientConn.SetWriteDeadline(base.TimeNow().Add(time.Second)),
	).IsNil()

	assert(
		serverConn.SetWriteDeadline(base.TimeNow().Add(time.Second)),
	).IsNil()

	_ = server.Close()
}
