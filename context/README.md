# context

一个轻量却强大的上下文管理器,一个请求会生成一个context，贯穿整个请求，context记录原始请求上下文，请求时间，客户端信息，权限校验信息，及负责判断是否内部调用，
及附带唯一traceId的日志记录器

支持http及fasthttp,并支持自定义的请求类型
![context](_assets/context.webp)

# usage
## Basic
### new
```go
import (
	stdcontext "context"
	"github.com/hopeio/context"
)

 ctx:= context.NewContext(stdcontext.Background())
```

### wrap and unwrap
```go
stdctx = ctx.Wrapper()
ctxi,_ = context.FromContext(stdctx)
```
## request context
### new
```go
import (
    "github.com/hopeio/context/httpctx"
)
func Handle(w http.ResponseWriter, r *http.Request) {
 ctx:= httpctx.FromRequest(httpctx.RequestCtx{r,w})
}

```
### wrap and unwrap
```go
func Middleware(w http.ResponseWriter, r *http.Request) {
    ctx:= httpctx.FromRequest(httpctx.RequestCtx{r,w})
    r.WithContext(ctx.Wrapper())
}
func Handle(w http.ResponseWriter, r *http.Request) {
    reqctx, ok := httpctx.FromContext(r.Context())
}

```

# auth 示例

```go
import (
  "errors"
  "encoding/json"
  "strings"
  "github.com/golang-jwt/jwt/v5"
  "github.com/hopeio/context/httpctx"
)
var tokenSecret = []byte("xxx")
type AuthInfo struct {
	Id uint64 `json:"id"`
}

type Authorization struct {
    *AuthInfo `json:"auth"`
    jwt.RegisteredClaims
    AuthInfoRaw string `json:"-"`
}

func (x *Authorization) Validate() error {
    return nil
}

func (x *Authorization) GenerateToken(secret []byte) (string, error) {
    tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, x)
    token, err := tokenClaims.SignedString(secret)
    return token, err
}

func (x *Authorization) ParseToken(token string, secret []byte) error {
    _, err := jwti.ParseToken(x, token, secret)
    if err != nil {
        return err
    }
    x.ID = x.AuthInfo.IdStr()
    authBytes, _ := json.Marshal(x.AuthInfo)
    x.AuthInfoRaw = string(authBytes)
    return nil
}

func GetAuth(ctx *httpctx.Context) (*AuthInfo, error) {
    signature := ctx.Token[strings.LastIndexByte(ctx.Token, '.')+1:]
	
    authorization := Authorization{AuthInfo: &AuthInfo{}}
    if err := authorization.ParseToken(ctx.Token, tokenSecret); err != nil {
        return nil, err
    }
    authInfo := authorization.AuthInfo
    ctx.AuthID = authInfo.IdStr()
    ctx.AuthInfo = authInfo
    ctx.AuthInfoRaw = authorization.AuthInfoRaw
	
    return authInfo, nil
}

```