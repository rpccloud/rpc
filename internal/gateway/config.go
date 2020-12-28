package gateway

import (
	"time"
)

// Config ...
type Config struct {
	numOfChannels    int
	transLimit       int
	heartbeat        time.Duration
	heartbeatTimeout time.Duration

	serverMaxSessions     int
	serverReadBufferSize  int
	serverWriteBufferSize int
	serverCacheTimeout    time.Duration

	clientRequestTimeout  time.Duration
	clientRequestInterval time.Duration
}

// GetDefaultConfig ...
func GetDefaultConfig() *Config {
	return &Config{
		numOfChannels:    48,
		transLimit:       4 * 1024 * 1024,
		heartbeat:        5 * time.Second,
		heartbeatTimeout: 8 * time.Second,

		serverMaxSessions:     10240000,
		serverReadBufferSize:  1200,
		serverWriteBufferSize: 1200,
		serverCacheTimeout:    12 * time.Second,

		clientRequestTimeout:  16 * time.Second,
		clientRequestInterval: 3 * time.Second,
	}
}

//
//// MaxSessions ...
//func (p *Config) MaxSessions() int {
//    return p.maxSessions
//}
//
//// NumOfChannels ...
//func (p *Config) NumOfChannels() int {
//    return p.numOfChannels
//}
//
//// TransLimit ...
//func (p *Config) TransLimit() int {
//    return p.transLimit
//}
//
//// ReadTimeout ...
//func (p *Config) ReadTimeout() time.Duration {
//    return p.readTimeout
//}
//
//// WriteTimeout ...
//func (p *Config) WriteTimeout() time.Duration {
//    return p.writeTimeout
//}
//
//// Heartbeat ...
//func (p *Config) Heartbeat() time.Duration {
//    return p.heartbeat
//}
//
//// CacheTimeout ...
//func (p *Config) CacheTimeout() time.Duration {
//    return p.cacheTimeout
//}
//
//// RequestTimeout ...
//func (p *Config) RequestTimeout() time.Duration {
//    return p.requestTimeout
//}
//
//// RequestInterval ...
//func (p *Config) RequestInterval() time.Duration {
//    return p.requestInterval
//}

//func ReadSessionConfig(
//    stream *core.Stream,
//) (SessionConfig, *base.Error) {
//    if numOfChannels, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if numOfChannels <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else if transLimit, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if transLimit <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else if readTimeout, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if readTimeout <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else if writeTimeout, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if writeTimeout <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else if heartbeat, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if heartbeat <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else if cacheTimeout, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if cacheTimeout <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else if requestTimeout, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if requestTimeout <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else if requestInterval, err := stream.ReadInt64(); err != nil {
//        return GetDefaultSessionConfig(), err
//    } else if requestInterval <= 0 {
//        return GetDefaultSessionConfig(), errors.ErrStream
//    } else {
//        return SessionConfig{
//            numOfChannels:   numOfChannels,
//            transLimit:      transLimit,
//            readTimeout:     time.Duration(readTimeout),
//            writeTimeout:    time.Duration(writeTimeout),
//            heartbeat:       time.Duration(heartbeat),
//            cacheTimeout:    time.Duration(cacheTimeout),
//            requestTimeout:  time.Duration(requestTimeout),
//            requestInterval: time.Duration(requestInterval),
//        }, nil
//    }
//}