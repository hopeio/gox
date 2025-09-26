package handlerwrap

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hopeio/gox/errors"
	httpx "github.com/hopeio/gox/net/http"
	"github.com/hopeio/gox/net/http/binding"
	"github.com/hopeio/gox/types"
)

type Service[REQ, RES any] func(ctx ReqResp, req REQ) (RES, *httpx.ErrRep)

type warpKey struct{}

var warpContextKey = warpKey{}

func WarpContext(v any) context.Context {
	return context.WithValue(context.Background(), warpContextKey, v)
}

func UnWarpContext(ctx context.Context) any {
	return ctx.Value(warpContextKey)
}

type ReqResp struct {
	*http.Request
	http.ResponseWriter
}

func HandlerWrap[REQ, RES any](service Service[*REQ, *RES]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := new(REQ)
		err := binding.Bind(r, req)
		if err != nil {
			httpx.RespErrCodeMsg(w, errors.InvalidArgument, err.Error())
			return
		}
		res, errRep := service(ReqResp{r, w}, req)
		if err != nil {
			errRep.Response(w)
			return
		}
		anyres := any(res)
		if httpres, ok := anyres.(httpx.ICommonResponseTo); ok {
			httpres.CommonResponse(httpx.CommonResponseWriter{ResponseWriter: w})
			return
		}
		if httpres, ok := anyres.(httpx.IHttpResponseTo); ok {
			httpres.Response(w)
			return
		}
		json.NewEncoder(w).Encode(res)
	})
}
func HandlerWrapCompatibleGRPC[REQ, RES any](method types.GrpcService[*REQ, *RES]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := new(REQ)
		err := binding.Bind(r, req)
		if err != nil {
			httpx.RespSuccessData(w, errors.InvalidArgument.Wrap(err))
			return
		}
		res, err := method(WarpContext(ReqResp{r, w}), req)
		if err != nil {
			httpx.ErrRepFrom(err).Response(w)
			return
		}
		anyres := any(res)
		if httpres, ok := anyres.(httpx.ICommonResponseTo); ok {
			httpres.CommonResponse(httpx.CommonResponseWriter{w})
			return
		}
		if httpres, ok := anyres.(httpx.IHttpResponseTo); ok {
			httpres.Response(w)
			return
		}
		w.Header().Set(httpx.HeaderContentType, httpx.ContentTypeJsonUtf8)
		json.NewEncoder(w).Encode(res)
	})
}
