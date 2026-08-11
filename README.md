# gox

[![Go Reference](https://pkg.go.dev/badge/github.com/hopeio/gox.svg)](https://pkg.go.dev/github.com/hopeio/gox)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

High-performance extensions around the Go standard library — import only what you need.

Go 标准库的高性能扩展——按子包按需引用。

```bash
go get github.com/hopeio/gox@latest
```

---

## English

### What is gox?

**gox** is a modular toolkit that fills gaps left by the Go standard library: zero-copy string conversions, pluggable JSON backends, Snowflake IDs, a backpressured task engine, rich containers, HTTP helpers, logging, generics-friendly slices/streams, and more.

It is not a framework. There is no central “app” object. Each subdirectory is a focused package you can depend on alone.

### Highlights

- **Zero-copy & low allocation** — `strings.ToBytes` / `FromBytes`, `unsafex.Cast` when you control lifetimes
- **Pluggable JSON** — default `encoding/json`; switch to sonic or go-json via build tags
- **IDs** — Snowflake, unique / ordered IDs, Base32/58/62/64 encodings
- **Scheduler** — worker pool with retry, child tasks, rate limits, and pending-queue backpressure
- **HTTP** — complete header / content-type constants and a small fluent client
- **Data** — GORM helpers (pagination, filter expressions, CRUD), Redis / ES helpers
- **Collections & streams** — LRU/LFU/ARC caches, heaps, sets; `iter.Stream` and generic `slices`
- **Everyday utilities** — zap-based logging, struct-tag flags with env defaults, `Option` / `Result` types, crypto, time, sync pools

### Quick start

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

### Package map

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

### Design rules

1. **Opt-in dependencies** — pay only for what you import.
2. **Performance first** — avoid pointless clones on hot paths.
3. **Stdlib-shaped APIs** — easy to learn, easy to drop in.
4. **Tests as docs** — see `*_test.go` next to each package.

### License

[MIT](LICENSE). Some vendored/subtree files may carry additional notices (see `LICENSE-Apache` and package-local LICENSE files).

---

## 中文

### gox 是什么？

**gox** 是围绕 Go 标准库补齐能力的模块化工具集：零拷贝字符串转换、可切换 JSON 后端、Snowflake ID、带背压的任务引擎、容器结构、HTTP 工具、日志、泛型切片/流式处理等。

它不是框架，没有统一的「应用入口」。每个子目录都是可独立依赖的包。

### 亮点

- **零拷贝 / 低分配** — `strings.ToBytes` / `FromBytes`、`unsafex.Cast`（需自行保证生命周期）
- **JSON 可切换** — 默认 `encoding/json`；通过 build tag 使用 sonic 或 go-json
- **ID** — Snowflake、唯一/有序 ID，以及多种进制编码
- **调度器** — worker 池，支持重试、子任务、限速与 pending 背压
- **HTTP** — Header / Content-Type 常量全集与轻量 Client
- **数据访问** — GORM 分页/过滤/CRUD，以及 Redis、ES 辅助
- **集合与流** — LRU/LFU/ARC、堆、集合；`iter.Stream` 与泛型 `slices`
- **日常工具** — zap 日志封装、结构体 tag 绑定 flag/env、`Option` / `Result`、加解密、时间、并发原语等

### 快速开始

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
	b := strings.ToBytes("hello") // 无拷贝视图；字符串存活期间勿改写底层字节
	fmt.Println(gox.TernaryOperator(id > 0, "ok", "fail"), len(b))
}
```

### 包一览

| 包 | 用途 |
|----|------|
| 根包 `gox` | 泛型微工具：`TernaryOperator`、`Pointer`、`Zero` |
| `strings`、`unsafe` | bytes↔string、无分配转换 |
| `encoding/json` | 可切换后端的 JSON |
| `idgen` | Snowflake 与 ID 编码 |
| `scheduler` | 并发任务引擎 |
| `net/http` | 常量、Client、中间件片段 |
| `log` | 面向生产的 zap 封装 |
| `flag` | 结构体 tag 绑定 flag/env |
| `iter`、`slices` | Stream 算子、Map/Filter/Reduce |
| `container` | 缓存、bitmap、一致性哈希、队列等 |
| `database` | SQL/GORM、Redis、Elasticsearch |
| `types` | `Option[T]`、`Result[T]` 等 |
| `crypto`、`sync`、`time`、`math`、`os`、`runtime`、`media`、`sdk`… | 各领域辅助 |

请 **按子包路径 import**，避免把用不到的依赖树拉进工程。

### 设计原则

1. **按需依赖** — 只为用到的包付成本。
2. **性能优先** — 热点路径避免无意义拷贝。
3. **贴近标准库** — API 好学、好替换。
4. **测试即文档** — 看各包旁的 `*_test.go`。

### 许可证

[MIT](LICENSE)。部分子树可能另有许可证说明（见 `LICENSE-Apache` 及包内 LICENSE）。
