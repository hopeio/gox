# gox

[![Go Reference](https://pkg.go.dev/badge/github.com/hopeio/gox.svg)](https://pkg.go.dev/github.com/hopeio/gox)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[English](README.md)

**Go 标准库的「扩展电池」** — 并发任务引擎、ID、可切换 JSON、零拷贝字符串、HTTP Client、缓存、Stream、GORM 辅助等。按子包按需引用。

```bash
go get github.com/hopeio/gox@latest
```

## gox 是什么？

**gox** 是面向生产服务的大型模块化工具库。它**不是**应用框架：没有统一入口。每个顶层目录都是独立包，API 风格贴近标准库。

常见场景：

| 领域 | 能力 |
|------|------|
| 工作负载 | 带优先级的任务引擎：重试、子任务、限速、pending 背压 |
| 标识 | 可配位宽 Snowflake、密码学 UniqueID、Base32/58/62/64 |
| 编码 | JSON（stdlib / sonic / go-json，build tag 切换）、Excel、msgpack、binary、base58/62 |
| 性能 | `string`↔`[]byte` 零拷贝、unsafe Cast、对象池、原子类型 |
| 网络 | 流畅 HTTP Client（重试、编解码、上下传）、Header/Content-Type 常量 |
| 数据 | GORM 分页与反射条件、Redis / Elasticsearch 辅助 |
| 集合 | LRU / LFU / ARC 缓存、堆、集合、bitmap、一致性哈希、队列 |
| 流式 | 基于 Go 1.23 `iter.Seq` 的链式算子；泛型切片 Map/Filter/Reduce |
| 运维向 | zap 日志（含 OTel）、结构体 tag 绑定 flag/env、校验、定时器、媒体 SDK |

## 亮点

- **`scheduler`** — 泛型 `Engine[KEY]`：worker 池、优先堆、子任务、`Limiter` / `KindLimiter`、`WithMaxPending` 背压、错误处理
- **`idgen`** — `NewSnowflake(node, nodeBits)`、`UniqueID()`、多进制字符串
- **`encoding/json`** — 即用 `Marshal` / `Unmarshal` / `MarshalToString`；`-tags sonic` 或 `-tags go_json`
- **`strings` / `unsafe`** — 生命周期可控时的无分配转换
- **`net/http/client`** — 流畅 `Client` / `Request`、自动编解码、重试、上传下载
- **`container/cache`** — Builder：`.LRU()` / `.LFU()` / `.ARC()`，支持 TTL
- **`iter` + `slices`** — `iter.Seq` 上的 Stream；谓词在前的 `Filter`
- **`database/sql/gorm`** — `FindList[T]`、`ConditionsBy` 反射条件
- **`flag` / `log` / `types`** — tag 驱动 CLI+环境变量、zap 封装、`Option` / `Result`

## 示例

### 并发任务引擎（`scheduler`）

```go
package main

import (
	"context"
	"fmt"

	"github.com/hopeio/gox/scheduler"
	"golang.org/x/time/rate"
)

func main() {
	eng := scheduler.NewEngine[int](12, scheduler.WithMaxPending[int](1024))
	eng.ErrHandlerUtilSuccess()
	eng.Limiter(rate.Limit(50), 50)

	eng.AddTasks(&scheduler.Task[int]{
		Key: 1,
		Run: func(ctx context.Context) ([]*scheduler.Task[int], error) {
			fmt.Println("parent")
			// 返回子任务即可扇出
			return []*scheduler.Task[int]{{
				Key: 2,
				Run: func(ctx context.Context) ([]*scheduler.Task[int], error) {
					fmt.Println("child")
					return nil, nil
				},
			}}, nil
		},
	})
	eng.Run()
}
```

### ID（`idgen`）

```go
import "github.com/hopeio/gox/idgen"

sf := idgen.NewSnowflake(1, 10) // node, nodeBits（stepBits = 22 - nodeBits）
id := sf.Generate()

u := idgen.UniqueID()
_ = u.Base58() // 另有 Hex / Base32 / Base62 / Base64
```

### 流畅 HTTP Client（`net/http/client`）

```go
import (
	"time"

	"github.com/hopeio/gox/net/http/client"
)

c := client.New().Timeout(5 * time.Second).RetryTimes(2).DisableLog()

var out map[string]any
_ = c.Get("https://httpbin.org/get", nil, &out)

_ = client.PostRequest("https://httpbin.org/post").
	ContentType(client.ContentTypeJson).
	Do(map[string]string{"hello": "gox"}, &out)
```

### 基于 `iter.Seq` 的 Stream（`iter`）

```go
import (
	"slices"

	"github.com/hopeio/gox/iter"
)

out := iter.StreamOf(slices.Values([]int{1, 2, 3, 4, 5})).
	Filter(func(n int) bool { return n%2 == 0 }).
	Map(func(n int) int { return n * 10 }).
	Collect() // [20, 40]
```

### 可切换 JSON（`encoding/json`）

```go
import jsonx "github.com/hopeio/gox/encoding/json"

s, _ := jsonx.MarshalToString(map[string]int{"a": 1})
var m map[string]int
_ = jsonx.UnmarshalFromString(s, &m)

// 更快后端（API 不变）：
//   go build -tags sonic     # amd64 + bytedance/sonic
//   go build -tags go_json   # goccy/go-json
```

### 缓存（`container/cache`）

```go
import (
	"time"

	"github.com/hopeio/gox/container/cache"
)

c := cache.New(1024).Expiration(5 * time.Minute).LRU() // 或 .LFU() / .ARC()
_ = c.Set("user:1", profile, cache.DefaultExpiration)
v, err := c.Get("user:1")
_, _ = v, err
```

### 零拷贝字符串（`strings`）

```go
import "github.com/hopeio/gox/strings"

b := strings.ToBytes("payload") // 共享底层存储；勿改写
s := strings.FromBytes(b)
_, _ = b, s
```

### GORM 列表查询（`database/sql/gorm`）

```go
import (
	sqlx "github.com/hopeio/gox/database/sql"
	gormx "github.com/hopeio/gox/database/sql/gorm"
)

list := &sqlx.List{Pagination: sqlx.Pagination{No: 1, Size: 20}}
rows, total, err := gormx.FindList[User](db, list)
// 反射条件：db.Clauses(gormx.ConditionsBy(&filter)...).Find(...)
_, _, _ = rows, total, err
```

### Flag + 环境变量（`flag`）与日志（`log`）

```go
import (
	"os"

	"github.com/hopeio/gox/flag"
	"github.com/hopeio/gox/log"
	"go.uber.org/zap"
)

type Cfg struct {
	Port int    `flag:"name:port;short:p;env:PORT;usage:listen port"`
	Host string `flag:"name:host;env:HOST;usage:bind host"`
}

var cfg Cfg
_ = flag.Bind(os.Args, &cfg)

log.SetDefaultLogger(log.NewProductionConfig("myapp").NewLogger())
log.Infow("listening", zap.Int("port", cfg.Port))
```

## 包一览

| 包 | 用途 |
|----|------|
| `scheduler` | 并发任务引擎（优先级、重试、子任务、限速、背压） |
| `idgen` | Snowflake、UniqueID / OrderedID / RandomID、进制编码 |
| `encoding` | JSON（可插拔）、Excel、binary、base58/62、msgpack、protobuf 辅助等 |
| `strings` / `unsafe` | 零拷贝转换、`Cast` / `CastSlice` |
| `net` | HTTP 常量、流畅 Client、URL/IP/mail |
| `container` | 缓存（LRU/LFU/ARC）、堆、集合、bitmap、一致性哈希、list/queue/stack/tree |
| `iter` / `slices` / `maps` | Stream 算子；泛型 Map/Filter/Reduce；map 辅助 |
| `database` | SQL/GORM 分页与条件；Redis；Elasticsearch |
| `log` | 基于 zap 的日志（OTel / slog 桥） |
| `flag` | 结构体 tag 绑定 pflag + 环境变量 |
| `types` | `Option[T]`、`Result[T]`、枚举、约束、请求/响应形态 |
| `sync` | singleflight、带锁容器、池、原子浮点 |
| `crypto` | AES、MD5、TLS 辅助 |
| `time` / `math` / `io` / `text` | 解析、定时器、decimal/geom、读写扩展、编码与模板 |
| `reflect` / `runtime` / `structtag` / `kvstruct` | 快速反射、goid/pprof、tag 解析、map↔struct |
| `media` / `sdk` / `mock` / `validator` / `terminal` | 图像视频、第三方薄封装、假数据、校验、进度条 |
| `os` / `archive` / `cmp` / `tools` | 文件系统/exec、zip、比较器、小工具（ddns、proxy…） |
| 根包 `gox` | 微型泛型（`TernaryOperator`、`Pointer`、`Zero`） |

请始终 **按子包路径 import**（例如 `github.com/hopeio/gox/scheduler`），避免把用不到的依赖树拉进工程。

## 设计原则

1. **按需付费** — 只为 import 的包承担成本。
2. **性能优先** — 热点路径优先零拷贝与复用。
3. **贴近标准库** — API 好学、好替换。
4. **测试即文档** — 各包旁 `*_test.go` 有更多用法。

## 许可证

[MIT](LICENSE)。部分子树可能另有许可证说明（见 `LICENSE-Apache` 及包内 LICENSE）。
