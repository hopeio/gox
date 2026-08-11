# gox

[![Go Reference](https://pkg.go.dev/badge/github.com/hopeio/gox.svg)](https://pkg.go.dev/github.com/hopeio/gox)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[中文文档](README.zh-CN.md)

```bash
go get github.com/hopeio/gox@latest
```

A modular Go library of production-ready building blocks. Every top-level folder is its own package — pull in a task engine, an ID generator, an HTTP client, or a zero-copy string helper without dragging the rest.

## Why it exists

The Go standard library is excellent and incomplete for day-to-day services. Teams re-implement the same pieces: worker pools with backpressure, Snowflake IDs, JSON that can swap backends, LRU caches, GORM pagination, fluent HTTP clients, zap logging, flag+env binding.

**gox** collects those pieces as independent packages with familiar, stdlib-like APIs.

## What you can do

| Need | Package | Capability |
|------|---------|------------|
| Run many jobs safely | `scheduler` | Generic `Engine[KEY]`: workers, priority, child tasks, rate limits, pending queue backpressure |
| Generate IDs | `idgen` | Snowflake (`Generate`), UniqueID, Base32/58/62/64 |
| Faster / swap JSON | `encoding/json` | Same API; `-tags sonic` or `-tags go_json` |
| Avoid copies | `strings`, `unsafe` | `ToBytes` / `FromBytes`, `Cast` / `CastSlice` |
| Call HTTP APIs | `net/http/client` | Fluent client: timeout, retry, JSON body, upload/download |
| In-process cache | `container/cache` | Builder → LRU / LFU / ARC + TTL |
| Process sequences | `iter`, `slices` | Stream on `iter.Seq`; Map / Filter / Reduce |
| Page GORM queries | `database/sql/gorm` | `FindList[T]`, reflective `ConditionsBy` |
| Configure a process | `flag`, `log` | Struct tags → pflag+env; zap logger |

Plus: Redis/ES helpers, heaps/sets/bitmap/consistent-hash, crypto, time/math/media, sync primitives, validators, thin SDKs, and more — see the package table below.

## Code samples

### Task engine

```go
eng := scheduler.NewEngine[string](16, scheduler.WithMaxPending[string](2048))
eng.ErrHandlerUtilSuccess()
eng.Limiter(rate.Limit(100), 100)

eng.AddTasks(&scheduler.Task[string]{
	Key: "crawl:home",
	Run: func(ctx context.Context) ([]*scheduler.Task[string], error) {
		// return more tasks to fan out
		return nil, nil
	},
})
eng.Run()
```

### Snowflake & UniqueID

```go
node := idgen.NewSnowflake(1, 10) // node id, node bit width
snow := node.Generate()

uid := idgen.UniqueID()
short := uid.Base58()
```

### HTTP client

```go
cli := client.New().
	Timeout(5 * time.Second).
	RetryTimes(3).
	DisableLog()

var body map[string]any
err := cli.Get("https://example.com/api", nil, &body)

err = client.PostRequest("https://example.com/api").
	ContentType(client.ContentTypeJson).
	Do(map[string]string{"k": "v"}, &body)
```

### `iter.Seq` stream

```go
result := iter.StreamOf(slices.Values([]int{1, 2, 3, 4, 5, 6})).
	Filter(func(n int) bool { return n > 2 }).
	Map(func(n int) int { return n * n }).
	Collect()
```

### JSON (backend selectable)

```go
s, err := jsonx.MarshalToString(payload)
err = jsonx.UnmarshalFromString(s, &payload)

// go build -tags sonic    # bytedance/sonic on amd64
// go build -tags go_json  # goccy/go-json
```

### Cache

```go
store := cache.New(4096).Expiration(10 * time.Minute).LRU()
_ = store.Set("sess:"+id, sess, cache.DefaultExpiration)
val, err := store.Get("sess:" + id)
```

### Zero-copy bytes view

```go
raw := stringsx.ToBytes(msg)  // no allocation; do not mutate
msg2 := stringsx.FromBytes(raw)
```

### GORM paging

```go
page := &sqlx.List{Pagination: sqlx.Pagination{No: 1, Size: 50}}
items, total, err := gormx.FindList[Order](db, page)
```

### Flags, env, logs

```go
type Options struct {
	Addr string `flag:"name:addr;short:a;env:ADDR;usage:listen address"`
}
var opt Options
_ = flag.Bind(os.Args, &opt)

log.SetDefaultLogger(log.NewProductionConfig("svc").NewLogger())
log.Infow("up", zap.String("addr", opt.Addr))
```

*(Imports omitted for brevity — use `github.com/hopeio/gox/<package>` paths.)*

## All top-level packages

| Path | Role |
|------|------|
| `scheduler` | Concurrent task engine |
| `idgen` | Distributed / unique IDs |
| `encoding` | JSON, Excel, binary, msgpack, base58/62, … |
| `strings` / `unsafe` | Zero-copy and unsafe views |
| `net` | HTTP constants & client, URL/IP/mail |
| `container` | Caches and classic data structures |
| `iter` / `slices` / `maps` | Streams and generic collection helpers |
| `database` | GORM/SQL, Redis, Elasticsearch |
| `log` | zap wrapper |
| `flag` | Struct → CLI flags + env |
| `types` | `Option` / `Result` and shared shapes |
| `sync` / `crypto` / `time` / `math` / `io` / `text` | Concurrency, crypto, time, numerics, I/O, text |
| `reflect` / `runtime` / `structtag` / `kvstruct` | Reflection and conversion utilities |
| `media` / `sdk` / `mock` / `validator` / `terminal` | Media, integrations, fakes, validation, TTY |
| `os` / `archive` / `cmp` / `tools` | OS helpers, zip, comparators, small tools |
| `.` (`gox`) | Tiny generics (`Pointer`, `Zero`, …) |

Import subpackages explicitly, e.g. `github.com/hopeio/gox/scheduler`.

## License

[MIT](LICENSE). Check subtree LICENSE files where present.
