# RPC
[![Lint](https://github.com/rpccloud/rpc/workflows/Lint/badge.svg)](https://github.com/rpccloud/rpc/actions?query=workflow%3ALint)
[![Test](https://github.com/rpccloud/rpc/workflows/Test/badge.svg)](https://github.com/rpccloud/rpc/actions?query=workflow%3ATest)
[![codecov](https://codecov.io/gh/rpccloud/rpc/branch/master/graph/badge.svg)](https://codecov.io/gh/rpccloud/rpc)
[![Go Report Card](https://goreportcard.com/badge/github.com/rpccloud/rpc)](https://goreportcard.com/report/github.com/rpccloud/rpc)
[![CodeFactor](https://www.codefactor.io/repository/github/rpccloud/rpc/badge)](https://www.codefactor.io/repository/github/rpccloud/rpc)

## 开发相关
#### 性能测试
```go
import _ "net/http/pprof"


go func() {
	log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

```bash
$ sudo brew install graphviz
$ curl "http://localhost:6060/debug/pprof/profile?seconds=10" -o cpu.prof
$ go tool pprof -web cpu.prof
```

#### Problems
1 How to avoid session flood <br>

#### 参考
https://github.com/tslearn/rpc-cluster-go/tree/984c4a17ffd777b268a13c0506aef83d9ba6b15d <br>
ssl 安全测试 https://www.ssllabs.com/ssltest/ <br>
