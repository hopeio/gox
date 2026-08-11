# gox

[![Go Reference](https://pkg.go.dev/badge/github.com/hopeio/gox.svg)](https://pkg.go.dev/github.com/hopeio/gox)
[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

[English](README.md)

```bash
go get github.com/hopeio/gox@latest
```

模块化的 Go 工具库。每个顶层目录都是独立包——你可以只引入任务引擎、ID、HTTP Client 或零拷贝字符串，而不必带上整库。

## 为什么需要它

标准库很强，但业务服务里仍要反复造轮子：带背压的 worker、Snowflake、可换后端的 JSON、LRU、GORM 分页、流畅 HTTP Client、zap 日志、flag+环境变量。

**gox** 把这些能力拆成风格接近标准库的独立包。

## 能做什么

| 需求 | 包 | 能力 |
|------|----|------|
| 安全跑大批任务 | `scheduler` | 泛型 `Engine[KEY]`：worker、优先级、子任务、限速、pending 背压 |
| 发号 | `idgen` | Snowflake（`Generate`）、UniqueID、Base32/58/62/64 |
| 更快 / 可换 JSON | `encoding/json` | API 不变；`-tags sonic` 或 `-tags go_json` |
| 少拷贝 | `strings`、`unsafe` | `ToBytes` / `FromBytes`、`Cast` |
| 调 HTTP | `net/http/client` | 超时、重试、JSON、上传下载 |
| 进程内缓存 | `container/cache` | Builder → LRU / LFU / ARC + TTL |
| 处理序列 | `iter`、`slices` | `iter.Seq` Stream；Map / Filter / Reduce |
| GORM 分页 | `database/sql/gorm` | `FindList[T]`、`ConditionsBy` |
| 进程配置 | `flag`、`log` | 结构体 tag → pflag+env；zap |

另有 Redis/ES、堆/集合/bitmap/一致性哈希、加解密、时间数学媒体、并发原语、校验器、薄 SDK 等，见下方包表。

## 代码示例

### 任务引擎

```go
eng := scheduler.NewEngine[string](16, scheduler.WithMaxPending[string](2048))
eng.ErrHandlerUtilSuccess()
eng.Limiter(rate.Limit(100), 100)

eng.AddTasks(&scheduler.Task[string]{
	Key: "crawl:home",
	Run: func(ctx context.Context) ([]*scheduler.Task[string], error) {
		return nil, nil // 可返回子任务扇出
	},
})
eng.Run()
```

### Snowflake 与 UniqueID

```go
node := idgen.NewSnowflake(1, 10)
snow := node.Generate()

uid := idgen.UniqueID()
short := uid.Base58()
```

### HTTP Client

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

### `iter.Seq` 流

```go
result := iter.StreamOf(slices.Values([]int{1, 2, 3, 4, 5, 6})).
	Filter(func(n int) bool { return n > 2 }).
	Map(func(n int) int { return n * n }).
	Collect()
```

### JSON（可换后端）

```go
s, err := jsonx.MarshalToString(payload)
err = jsonx.UnmarshalFromString(s, &payload)

// go build -tags sonic
// go build -tags go_json
```

### 缓存

```go
store := cache.New(4096).Expiration(10 * time.Minute).LRU()
_ = store.Set("sess:"+id, sess, cache.DefaultExpiration)
val, err := store.Get("sess:" + id)
```

### 零拷贝

```go
raw := stringsx.ToBytes(msg) // 勿改写底层字节
msg2 := stringsx.FromBytes(raw)
```

### GORM 分页

```go
page := &sqlx.List{Pagination: sqlx.Pagination{No: 1, Size: 50}}
items, total, err := gormx.FindList[Order](db, page)
```

### Flag、环境变量、日志

```go
type Options struct {
	Addr string `flag:"name:addr;short:a;env:ADDR;usage:listen address"`
}
var opt Options
_ = flag.Bind(os.Args, &opt)

log.SetDefaultLogger(log.NewProductionConfig("svc").NewLogger())
log.Infow("up", zap.String("addr", opt.Addr))
```

（示例省略 import，路径形如 `github.com/hopeio/gox/<package>`。）

## 顶层包一览

| 路径 | 作用 |
|------|------|
| `scheduler` | 并发任务引擎 |
| `idgen` | 分布式 / 唯一 ID |
| `encoding` | JSON、Excel、binary、msgpack、base58/62… |
| `strings` / `unsafe` | 零拷贝与 unsafe 视图 |
| `net` | HTTP 常量与 Client、URL/IP/mail |
| `container` | 缓存与经典数据结构 |
| `iter` / `slices` / `maps` | 流与泛型集合 |
| `database` | GORM/SQL、Redis、ES |
| `log` | zap 封装 |
| `flag` | 结构体 → CLI + env |
| `types` | `Option` / `Result` 等 |
| `sync` / `crypto` / `time` / `math` / `io` / `text` | 并发、加密、时间、数值、I/O、文本 |
| `reflect` / `runtime` / `structtag` / `kvstruct` | 反射与转换 |
| `media` / `sdk` / `mock` / `validator` / `terminal` | 媒体、集成、假数据、校验、终端 |
| `os` / `archive` / `cmp` / `tools` | OS、zip、比较、小工具 |
| `.`（`gox`） | 微型泛型 |

请显式 import 子包，例如 `github.com/hopeio/gox/scheduler`。

## 许可证

[MIT](LICENSE)。部分子树另有 LICENSE 时请一并查看。
