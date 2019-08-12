package core

type fnCacheFunc = func(
	ctx Context,
	stream *rpcStream,
	fn interface{},
) bool

type fnProcessorCallback = func(
	stream *rpcStream,
	success bool,
)

// Any common Any type
type Any = interface{}

// Array common Array type
type Array = []interface{}

// Map common Map type
type Map = map[string]interface{}

var (
	readTypeArray Array
	readTypeMap   Map
)
