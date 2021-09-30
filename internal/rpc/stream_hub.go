package rpc

import (
	"github.com/rpccloud/rpc/internal/base"
)

type StreamHubCallback struct {
	OnRPCRequestStream        func(stream *Stream)
	OnRPCResponseOKStream     func(stream *Stream)
	OnRPCResponseErrorStream  func(stream *Stream)
	OnRPCBoardCastStream      func(stream *Stream)
	OnSystemErrorReportStream func(sessionID uint64, err *base.Error)
}

// StreamHub ...
type StreamHub struct {
	logger   *base.Logger
	logLevel base.ErrorLevel
	callback StreamHubCallback
}

// NewStreamHub ...
func NewStreamHub(
	isLogErrorToScreen bool,
	logFile string,
	logLevel base.ErrorLevel,
	callback StreamHubCallback,
) *StreamHub {
	logger, err := base.NewLogger(isLogErrorToScreen, logFile)

	ret := &StreamHub{
		logger:   logger,
		logLevel: logLevel,
		callback: callback,
	}

	if err != nil {
		ret.OnReceiveStream(MakeSystemErrorStream(err))
	}

	return ret
}

// OnReceiveStream ...
func (p *StreamHub) OnReceiveStream(stream *Stream) {
	if stream != nil {
        stream.SetReadPos(streamPosBody)
		fn := (func(stream *Stream))(nil)
		switch stream.GetKind() {
		case StreamKindRPCRequest:
			fn = p.callback.OnRPCRequestStream
		case StreamKindRPCResponseOK:
			fn = p.callback.OnRPCResponseOKStream
		case StreamKindRPCResponseError:
			fn = p.callback.OnRPCResponseErrorStream
		case StreamKindRPCBoardCast:
			fn = p.callback.OnRPCBoardCastStream
		case StreamKindSystemErrorReport:
			// err is definitely not nil
			_, err := ParseResponseStream(stream)

			if err.GetLevel()&p.logLevel == 0 {
				return
			}

			p.logger.Log(err.ReportString(
				stream.GetSourceID(),
				stream.GetSessionID(),
			))

			if p.callback.OnSystemErrorReportStream != nil {
				p.callback.OnSystemErrorReportStream(stream.GetSessionID(), err)
			}

			return
		}

		if fn != nil {
			fn(stream)
		}
	}
}

// Close ...
func (p *StreamHub) Close() bool {
	if err := p.logger.Close(); err != nil {
		errStream := MakeSystemErrorStream(err)
		p.OnReceiveStream(errStream)
		return false
	}

	return true
}
