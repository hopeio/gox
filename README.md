# gox

[![Go Reference](https://pkg.go.dev/badge/github.com/hopeio/gox.svg)](https://pkg.go.dev/github.com/hopeio/gox)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[中文文档](README.zh-CN.md)

High-performance extensions around the Go standard library — import only what you need.

```bash
go get github.com/hopeio/gox@latest
```

## What is gox?

**gox** is a modular toolkit that fills gaps left by the Go standard library: zero-copy string conversions, pluggable JSON backends, Snowflake IDs, a backpressured task engine, rich containers, HTTP helpers, logging, generics-friendly slices/streams, and more.

It is not a framework. There is no central “app” object. Each subdirectory is a focused package you can depend on alone.

## Highlights

- **Zero-copy & low allocation** — `strings.ToBytes` / `FromBytes`, `unsafex.Cast` when you control lifetimes
- **Pluggable JSON** — default `encoding/json`; switch to sonic or go-json via build tags
- **IDs** — Snowflake, unique / ordered IDs, Base32/58/62/64 encodings
- **Scheduler** — worker pool with retry, child tasks, rate limits, and pending-queue backpressure
- **HTTP** — complete header / content-type constants and a small fluent client
- **Data** — GORM helpers (pagination, filter expressions, CRUD), Redis / ES helpers
- **Collections & streams** — LRU/LFU/ARC caches, heaps, sets; `iter.Stream` and generic `slices`
- **Everyday utilities** — zap-based logging, struct-tag flags with env defaults, `Option` / `Result` types, crypto, time, sync pools

## Quick start

```go
package main

import (
	"fmt"

	"github.com/hopeio/gox"
	"github.com/hopeio/gox/idgen"
	"github.com/hopeio/gox/strings"
)

func main() {
	id := idgen.NewSnowflake(1).Next()
	b := strings.ToBytes("hello") // no copy; do not mutate while string lives
	fmt.Println(gox.TernaryOperator(id > 0, "ok", "fail"), len(b))
}
```

## Package map

| Package | Purpose |
|---------|---------|
| root (`gox`) | Tiny generics: `TernaryOperator`, `Pointer`, `Zero` |
| `strings`, `unsafe` | Bytes↔string, allocation-free casts |
| `encoding/json` | JSON with selectable backend |
| `idgen` | Snowflake and ID encodings |
| `scheduler` | Concurrent task engine |
| `net/http` | Constants, client, middleware bits |
| `log` | Production-oriented zap wrapper |
| `flag` | Bind flags/env from struct tags |
| `iter`, `slices` | Stream operators, Map/Filter/Reduce |
| `container` | Caches, bitmap, hash ring, queues… |
| `database` | SQL/GORM, Redis, Elasticsearch |
| `types` | `Option[T]`, `Result[T]`, … |
| `crypto`, `sync`, `time`, `math`, `os`, `runtime`, `media`, `sdk`, … | Domain helpers |

Prefer importing a **subpackage path** so unused trees stay out of your binary and module graph.

## Design rules

1. **Opt-in dependencies** — pay only for what you import.
2. **Performance first** — avoid pointless clones on hot paths.
3. **Stdlib-shaped APIs** — easy to learn, easy to drop in.
4. **Tests as docs** — see `*_test.go` next to each package.

## License

[MIT](LICENSE). Some vendored/subtree files may carry additional notices (see `LICENSE-Apache` and package-local LICENSE files).
