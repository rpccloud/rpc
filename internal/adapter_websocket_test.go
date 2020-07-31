package internal

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync/atomic"
	"testing"
	"time"
)

func testWithStreamConn(
	runOnServer func(IAdapter, IStreamConn),
	runOnClient func(IAdapter, IStreamConn),
) []Error {
	ret := make([]Error, 0)
	lock := NewLock()
	fnOnError := func(err Error) {
		lock.DoWithLock(func() {
			ret = append(ret, err)
		})
	}

	waitCH := make(chan bool)
	serverAdapter := NewWebSocketServerAdapter("127.0.0.1:12345")
	clientAdapter := NewWebSocketClientEndPoint("ws://127.0.0.1:12345")
	go func() {
		serverAdapter.Open(func(conn IStreamConn) {
			runOnServer(serverAdapter, conn)
		}, fnOnError)
		waitCH <- true
	}()

	go func() {
		time.Sleep(200 * time.Millisecond)
		clientAdapter.Open(func(conn IStreamConn) {
			runOnClient(clientAdapter, conn)
		}, fnOnError)
		time.Sleep(time.Second)
		serverAdapter.Close(fnOnError)
		waitCH <- true
	}()

	for i := 0; i < 2; i++ {
		<-waitCH
	}

	return ret
}

func TestNewWebSocketStreamConn(t *testing.T) {
	assert := NewAssert(t)

	// Test(1)
	wsConn := &websocket.Conn{}
	sc1 := newWebSocketStreamConn(wsConn)
	assert(fmt.Sprintf("%p", wsConn.CloseHandler())).
		Equals(fmt.Sprintf("%p", sc1.onCloseMessage))
	assert(sc1.status).Equals(webSocketStreamConnRunning)
	assert(sc1.conn).Equals(wsConn)
	assert(cap(sc1.canClose)).Equals(1)
}

func TestWebSocketStreamConn_onCloseMessage(t *testing.T) {
	assert := NewAssert(t)

	//// Test(1) webSocketStreamConnRunning => webSocketStreamConnCanClose
	//assert(testWithStreamConn(
	//  func(server IAdapter, conn IStreamConn) {
	//    testConn := conn.(*webSocketStreamConn)
	//    assert(atomic.LoadInt32(&testConn.status)).
	//      Equals(webSocketStreamConnRunning)
	//    _, _ = testConn.ReadStream(10*time.Second, 999999)
	//    assert(atomic.LoadInt32(&testConn.status)).
	//      Equals(webSocketStreamConnCanClose)
	//  },
	//  func(client IAdapter, conn IStreamConn) {
	//    testConn := conn.(*webSocketStreamConn)
	//    assert(atomic.LoadInt32(&testConn.status)).
	//      Equals(webSocketStreamConnRunning)
	//    assert(testConn.onCloseMessage(
	//      websocket.CloseNormalClosure,
	//      "",
	//    )).IsNil()
	//    assert(atomic.LoadInt32(&testConn.status)).
	//      Equals(webSocketStreamConnCanClose)
	//    time.Sleep(300 * time.Millisecond)
	//  },
	//)).Equals([]Error{})

	// Test(2) webSocketStreamConnClosing => webSocketStreamConnCanClose
	assert(testWithStreamConn(
		func(server IAdapter, conn IStreamConn) {
			testConn := conn.(*webSocketStreamConn)
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnRunning)
			_, _ = testConn.ReadStream(10*time.Second, 999999)
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnCanClose)
		},
		func(client IAdapter, conn IStreamConn) {
			go func() {
				for {
					_, err := conn.ReadStream(time.Second, 999999)
					if err != nil {
						return
					}
				}
			}()
			testConn := conn.(*webSocketStreamConn)
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnRunning)
			assert(testConn.Close()).IsNil()
			assert(atomic.LoadInt32(&testConn.status)).
				Equals(webSocketStreamConnClosed)
			time.Sleep(300 * time.Millisecond)
		},
	)).Equals([]Error{})
}

//
//func TestNewWebSocketServerEndPoint(t *testing.T) {
//	assert := internal.NewAssert(t)
//
//	server := NewWebSocketServerAdapter("127.0.0.1:20080", "/test")
//
//	for i := 0; i < 2; i++ {
//		assert(
//			server.Open(func(conn IStreamConn) {
//				fmt.Println(conn)
//			}, func(err Error) {
//				fmt.Println(err)
//			}),
//		).IsTrue()
//
//		time.Sleep(2 * time.Second)
//
//		assert(
//			server.Close(func(err Error) {
//				fmt.Println(err)
//			}),
//		).IsTrue()
//	}
//}
//
//func TestNewWebSocketClientEndPoint(t *testing.T) {
//	assert := internal.NewAssert(t)
//	server := NewWebSocketServerAdapter("127.0.0.1:20080", "/test")
//	server.Open(func(conn IStreamConn) {
//		time.Sleep(3 * time.Second)
//	}, func(err Error) {
//		fmt.Println(err)
//	})
//
//	client := NewWebSocketClientEndPoint("ws://127.0.0.1:20080/test")
//	assert(
//		client.Open(func(conn IStreamConn) {
//			time.Sleep(3 * time.Second)
//		}, func(err Error) {
//			fmt.Println(err)
//		}),
//	).IsTrue()
//
//	time.Sleep(1 * time.Second)
//
//	assert(
//		client.Close(func(err Error) {
//			fmt.Println(err)
//		}),
//	).IsTrue()
//
//	server.Close(func(err Error) {
//		fmt.Println(err)
//	})
//}