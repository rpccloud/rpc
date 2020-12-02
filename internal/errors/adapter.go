package errors

import (
	"github.com/rpccloud/rpc/internal/base"
)

const streamConnErrorSeg = 2000

var (
	// ErrStreamConnIsClosed ... *
	ErrStreamConnIsClosed = base.DefineNetError(
		(streamConnErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"stream conn is closed",
	)
)

const websocketStreamConnErrorSeg = 2001

var (
	// ErrWebsocketStreamConnWSConnWriteMessage ... *
	ErrWebsocketStreamConnWSConnWriteMessage = base.DefineNetError(
		(websocketStreamConnErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"",
	)

	// ErrWebsocketStreamConnWSConnSetReadDeadline ... *
	ErrWebsocketStreamConnWSConnSetReadDeadline = base.DefineNetError(
		(websocketStreamConnErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"",
	)

	// ErrWebsocketStreamConnWSConnReadMessage ... *
	ErrWebsocketStreamConnWSConnReadMessage = base.DefineNetError(
		(websocketStreamConnErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	// ErrWebsocketStreamConnDataIsNotBinary ... *
	ErrWebsocketStreamConnDataIsNotBinary = base.DefineSecurityError(
		(websocketStreamConnErrorSeg<<16)|4,
		base.ErrorLevelWarn,
		"websocket data is not binary",
	)

	// ErrWebsocketStreamConnStreamIsNil ... *
	ErrWebsocketStreamConnStreamIsNil = base.DefineKernelError(
		(websocketStreamConnErrorSeg<<16)|5,
		base.ErrorLevelFatal,
		"stream is nil",
	)

	// ErrWebsocketStreamConnWSConnClose ... *
	ErrWebsocketStreamConnWSConnClose = base.DefineNetError(
		(websocketStreamConnErrorSeg<<16)|6,
		base.ErrorLevelWarn,
		"",
	)
)

const websocketServerAdapterErrorSeg = 2002

var (
	// ErrWebsocketServerAdapterUpgrade ... *
	ErrWebsocketServerAdapterUpgrade = base.DefineSecurityError(
		(websocketServerAdapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"websocket upgrade error",
	)

	// ErrWebsocketServerAdapterWSServerListenAndServe ... *
	ErrWebsocketServerAdapterWSServerListenAndServe = base.DefineConfigError(
		(websocketServerAdapterErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"",
	)

	// ErrWebsocketServerAdapterAlreadyRunning ... *
	ErrWebsocketServerAdapterAlreadyRunning = base.DefineKernelError(
		(websocketServerAdapterErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is already running",
	)

	// ErrWebsocketServerAdapterNotRunning ... *
	ErrWebsocketServerAdapterNotRunning = base.DefineKernelError(
		(websocketServerAdapterErrorSeg<<16)|4,
		base.ErrorLevelFatal,
		"it is not running",
	)

	// ErrWebsocketServerAdapterWSServerClose ... *
	ErrWebsocketServerAdapterWSServerClose = base.DefineNetError(
		(websocketServerAdapterErrorSeg<<16)|5,
		base.ErrorLevelError,
		"",
	)

	// ErrWebsocketServerAdapterCloseTimeout ... *
	ErrWebsocketServerAdapterCloseTimeout = base.DefineKernelError(
		(websocketServerAdapterErrorSeg<<16)|6,
		base.ErrorLevelError,
		"close timeout",
	)
)

const websocketClientAdapterErrorSeg = 2003

var (
	// ErrWebsocketClientAdapterDial ... *
	ErrWebsocketClientAdapterDial = base.DefineConfigError(
		(websocketClientAdapterErrorSeg<<16)|1,
		base.ErrorLevelFatal,
		"",
	)

	// ErrWebsocketClientAdapterAlreadyRunning ... *
	ErrWebsocketClientAdapterAlreadyRunning = base.DefineKernelError(
		(websocketClientAdapterErrorSeg<<16)|2,
		base.ErrorLevelFatal,
		"it is already running",
	)

	// ErrWebsocketClientAdapterNotRunning ... *
	ErrWebsocketClientAdapterNotRunning = base.DefineKernelError(
		(websocketClientAdapterErrorSeg<<16)|3,
		base.ErrorLevelFatal,
		"it is not running",
	)

	// ErrWebsocketClientAdapterCloseTimeout ... *
	ErrWebsocketClientAdapterCloseTimeout = base.DefineKernelError(
		(websocketClientAdapterErrorSeg<<16)|4,
		base.ErrorLevelError,
		"close timeout",
	)
)

const tcpStreamConnErrorSeg = 2101

var (
	ErrTCPStreamConnReadStream = base.DefineNetError(
		(tcpStreamConnErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"",
	)
	ErrTCPStreamConnWriteStream = base.DefineNetError(
		(tcpStreamConnErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"",
	)
	ErrTCPStreamConnClose = base.DefineNetError(
		(tcpStreamConnErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)
)

const tcpServerAdapterErrorSeg = 2102

var (
	ErrTCPServerAdapterAlreadyRunning = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|1,
		base.ErrorLevelWarn,
		"it is already running",
	)

	ErrTCPServerAdapterNotRunning = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|2,
		base.ErrorLevelWarn,
		"it is not running",
	)

	ErrTCPServerAdapterListen = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)

	ErrTCPServerAdapterAccept = base.DefineNetError(
		(tcpServerAdapterErrorSeg<<16)|3,
		base.ErrorLevelWarn,
		"",
	)
)
