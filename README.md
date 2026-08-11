# gox

[![Go Reference](https://pkg.go.dev/badge/github.com/hopeio/gox.svg)](https://pkg.go.dev/github.com/hopeio/gox)
[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**Go standard library, upgraded.**  
高性能、可按需引用的 Go 工具集——零拷贝字符串、可切换 JSON 后端、Snowflake、带背压的任务引擎、GORM 分页与 HTTP 常量全集。

> 不是又一个「杂物抽屉」。每个子包都能独立成库；你只 `import` 真正用到的那一块。

```bash
go get github.com/hopeio/gox@latest
```

## 为什么选 gox

| 痛点 | gox 怎么做 |
|------|------------|
| `[]byte` ↔ `string` 频繁拷贝 | `strings.ToBytes` / `FromBytes`、`unsafex.Cast` 无分配转换 |
| JSON 库绑死 | build tag 在 `encoding/json` / [sonic](https://github.com/bytedance/sonic) / [go-json](https://github.com/goccy/go-json) 间切换 |
| 分布式 ID / 有序 ID 各写一套 | `idgen`：Snowflake、UniqueID、Base58/62 等 |
| 爬虫/批处理要限流、重试、子任务 | `scheduler.Engine`：worker 池 + 背压 + 统计 |
| HTTP Header / Content-Type 魔法字符串 | `net/http` 常量全集，含 gRPC / Device 等 |
| 切片/流式处理样板代码多 | 泛型 `slices` + `iter.Stream` |

## 快速上手

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
	raw := strings.ToBytes("hello") // 无拷贝视图（生命周期内勿改写）
	fmt.Println(gox.TernaryOperator(id > 0, "ok", "fail"), len(raw))
}
```

任务引擎：

```go
import "github.com/hopeio/gox/scheduler"

eng := scheduler.NewEngine(/* options */)
eng.AddTasks(/* ... */)
```

更多用法见各包 `*_test.go`（本仓库以测试即文档）。

## 包一览

| 包 | 能力 |
|----|------|
| `.` | `TernaryOperator` / `Pointer` / `Zero` 等泛型微工具 |
| [`strings`](strings/) · [`unsafe`](unsafe/) | 零拷贝 bytes↔string、无分配 Cast |
| [`encoding/json`](encoding/json/) | 可切换 JSON 后端；`MarshalToString` |
| [`idgen`](idgen/) | Snowflake / UniqueID / 多进制编码 |
| [`scheduler`](scheduler/) | 限流、重试、子任务、`maxPending` 背压 |
| [`net/http`](net/http/) | Header / Content-Type 常量、流畅 HTTP Client |
| [`log`](log/) | zap 开箱封装（OTel / slog / grpclog） |
| [`flag`](flag/) | 结构体 tag → pflag，支持 env / default |
| [`iter`](iter/) · [`slices`](slices/) | Stream 算子、泛型 Map/Filter/Reduce |
| [`container`](container/) | bitmap、LRU/LFU/ARC、consistent hash、heap… |
| [`database/sql/gorm`](database/sql/gorm/) | 分页/过滤表达式、CRUD、连接池指标钩子 |
| [`types`](types/) | `Option[T]` / `Result[T]` / Pair / Enum |
| [`crypto`](crypto/) · [`sync`](sync/) · [`time`](time/) · [`math`](math/) … | 按需扩展 |

完整目录以仓库为准；**按子路径 import**，避免一次拉进全部依赖。

## 设计原则

1. **按需依赖** — 子包独立；不要 `import` 整仓根以外的无关树。
2. **高性能优先** — 能零拷贝就不 clone；热点路径有 build tag 加速。
3. **贴近标准库** — API 风格像 `stdlib`，学习成本低。
4. **可组合** — 与 [initialize](https://github.com/hopeio/initialize) / [mix](https://github.com/hopeio/mix) / [protobuf](https://github.com/hopeio/protobuf) 分层协作，不互相绑架。

## hopeio 生态

```
protobuf ──► 生成 gRPC / HTTP / OpenAPI
    │
    ▼
  mix ──► 同端口 HTTP + gRPC 运行时
    │
    ▼
initialize ──► 配置 / DAO 反射注入
    │
    ▼
   gox ──► 日志、HTTP、ID、调度、编码……
```

| 仓库 | 定位 |
|------|------|
| **gox**（本仓库） | 标准库级工具与运行时积木 |
| [initialize](https://github.com/hopeio/initialize) | 配置中心 + DAO 自动注入 |
| [mix](https://github.com/hopeio/mix) | 同进程 HTTP/gRPC 微服务运行时 |
| [protobuf](https://github.com/hopeio/protobuf) | `protogen` 与公共 proto / 插件 |

## 文档与参考

- [pkg.go.dev/github.com/hopeio/gox](https://pkg.go.dev/github.com/hopeio/gox)
- 各子包单元测试即可运行示例

## License

[MIT](LICENSE) · Copyright © hopeio

部分子树（如 cache / 第三方薄封装）可能附带各自 LICENSE，使用时请一并核对。
