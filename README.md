# gox

[![Go Reference](https://pkg.go.dev/badge/github.com/hopeio/gox.svg)](https://pkg.go.dev/github.com/hopeio/gox)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[中文文档](README.zh-CN.md)

**Batteries for Go’s standard library** — concurrency engine, IDs, pluggable JSON, zero-copy strings, HTTP client, caches, streams, GORM helpers, and more. Import only the packages you use.

```bash
go get github.com/hopeio/gox@latest
```

## What is gox?

**gox** is a large, modular utility library for production Go services. It is **not** an application framework: there is no single entrypoint. Each top-level directory is an independent package with stdlib-shaped APIs.

Typical use cases:

| Area | What you get |
|------|----------------|
| Workloads | Priority task engine with retry, child tasks, rate limits, pending backpressure |
| Identity | Configurable Snowflake, crypto UniqueID, Base32/58/62/64 encodings |
| Encoding | JSON (stdlib / sonic / go-json via build tags), Excel, msgpack, binary, base58/62 |
| Performance | Zero-copy `string`↔`[]byte`, unsafe casts, sync pools, atomic helpers |
| Networking | Fluent HTTP client (retry, marshal, download/upload), header/content-type constants |
| Data | GORM pagination & reflective filters, Redis / Elasticsearch helpers |
| Collections | LRU / LFU / ARC caches, heaps, sets, bitmap, consistent hash, queues |
| Streams | Chainable operators on Go 1.23 `iter.Seq`, generic slice Map/Filter/Reduce |
| Ops | zap logging (+ OTel bridge), struct-tag flags/env, validators, timers, media SDKs |

## Highlights

- **`scheduler`** — generic `Engine[KEY]`: worker pool, priority heap, subtasks, `Limiter` / `KindLimiter`, `WithMaxPending` backpressure, error handlers
- **`idgen`** — `NewSnowflake(node, nodeBits)`, `UniqueID()`, multi-radix string forms
- **`encoding/json`** — drop-in `Marshal` / `Unmarshal` / `MarshalToString`; `-tags sonic` or `-tags go_json`
- **`strings` / `unsafe`** — allocation-free conversions when lifetimes allow
- **`net/http/client`** — fluent `Client` / `Request`, auto (un)marshal, retries, upload & download
- **`container/cache`** — builder API: `.LRU()` / `.LFU()` / `.ARC()` with TTL
- **`iter` + `slices`** — Stream pipeline over `iter.Seq`; predicate-first `Filter`
- **`database/sql/gorm`** — `FindList[T]`, `ConditionsBy` reflective clauses
- **`flag` / `log` / `types`** — tag-driven CLI+env, zap wrapper, `Option` / `Result`

## Examples

### Concurrent task engine (`scheduler`)

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
			// return child tasks to fan out work
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

### IDs (`idgen`)

```go
import "github.com/hopeio/gox/idgen"

sf := idgen.NewSnowflake(1, 10) // node, nodeBits (stepBits = 22 - nodeBits)
id := sf.Generate()

u := idgen.UniqueID()
_ = u.Base58() // also Hex / Base32 / Base62 / Base64
```

### Fluent HTTP client (`net/http/client`)

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

### Stream over `iter.Seq` (`iter`)

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

### Pluggable JSON (`encoding/json`)

```go
import jsonx "github.com/hopeio/gox/encoding/json"

s, _ := jsonx.MarshalToString(map[string]int{"a": 1})
var m map[string]int
_ = jsonx.UnmarshalFromString(s, &m)

// Faster backends (same API):
//   go build -tags sonic     # amd64 + bytedance/sonic
//   go build -tags go_json   # goccy/go-json
```

### Cache (`container/cache`)

```go
import (
	"time"

	"github.com/hopeio/gox/container/cache"
)

c := cache.New(1024).Expiration(5 * time.Minute).LRU() // or .LFU() / .ARC()
_ = c.Set("user:1", profile, cache.DefaultExpiration)
v, err := c.Get("user:1")
_, _ = v, err
```

### Zero-copy strings (`strings`)

```go
import "github.com/hopeio/gox/strings"

b := strings.ToBytes("payload") // shares backing store; do not mutate
s := strings.FromBytes(b)
_, _ = b, s
```

### GORM list query (`database/sql/gorm`)

```go
import (
	sqlx "github.com/hopeio/gox/database/sql"
	gormx "github.com/hopeio/gox/database/sql/gorm"
)

list := &sqlx.List{Pagination: sqlx.Pagination{No: 1, Size: 20}}
rows, total, err := gormx.FindList[User](db, list)
// reflective filters: db.Clauses(gormx.ConditionsBy(&filter)...).Find(...)
_, _, _ = rows, total, err
```

### Flags + env (`flag`) & logging (`log`)

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

## Package map

| Package | Purpose |
|---------|---------|
| `scheduler` | Concurrent task engine (priority, retry, subtasks, rate limit, backpressure) |
| `idgen` | Snowflake, UniqueID / OrderedID / RandomID, radix encodings |
| `encoding` | JSON (pluggable), Excel, binary, base58/62, msgpack, protobuf helpers, … |
| `strings` / `unsafe` | Zero-copy conversions, `Cast` / `CastSlice` |
| `net` | HTTP constants, fluent client, URL/IP/mail helpers |
| `container` | Cache (LRU/LFU/ARC), heap, set, bitmap, consistent hash, list/queue/stack/tree |
| `iter` / `slices` / `maps` | Stream operators; generic Map/Filter/Reduce; map helpers |
| `database` | SQL/GORM pagination & conditions; Redis; Elasticsearch |
| `log` | zap-based logger with OTel / slog bridges |
| `flag` | Struct-tag binding to pflag + environment variables |
| `types` | `Option[T]`, `Result[T]`, enums, constraints, request/response shapes |
| `sync` | Singleflight, locked containers, pools, atomic floats |
| `crypto` | AES, MD5, TLS helpers |
| `time` / `math` / `io` / `text` | Parsing, timers, decimal/geom, readers, encoding & templates |
| `reflect` / `runtime` / `structtag` / `kvstruct` | Fast reflection, goid/pprof, tag parsing, map↔struct |
| `media` / `sdk` / `mock` / `validator` / `terminal` | Image/video, third-party thin SDKs, fakes, validation, progress UI |
| `os` / `archive` / `cmp` / `tools` | Filesystem/exec, zip, comparators, small CLIs (ddns, proxy, …) |
| root `gox` | Tiny generics (`TernaryOperator`, `Pointer`, `Zero`) |

Always import a **subpackage path** (e.g. `github.com/hopeio/gox/scheduler`) so unused trees stay out of your module graph.

## Design rules

1. **Opt-in cost** — pay only for packages you import.
2. **Performance first** — prefer zero-copy and reuse on hot paths.
3. **Stdlib-shaped** — APIs should feel familiar and replaceable.
4. **Tests as docs** — see `*_test.go` beside each package for more patterns.

## License

[MIT](LICENSE). Some vendored/subtree files may carry additional notices (see `LICENSE-Apache` and package-local LICENSE files).
