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

type Service[REQ, RES any] func(ctx ReqResp, req REQ) (RES, *httpx.ErrResp)

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
			errRep.Respond(w)
			return
		}
		anyres := any(res)
		if httpres, ok := anyres.(httpx.ICommonRespond); ok {
			httpres.CommonRespond(httpx.CommonResponseWriter{ResponseWriter: w})
			return
		}
		if httpres, ok := anyres.(httpx.IRespond); ok {
			httpres.Respond(w)
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
			httpx.ErrRespFrom(err).Respond(w)
			return
		}
		anyres := any(res)
		if httpres, ok := anyres.(httpx.ICommonRespond); ok {
			httpres.CommonRespond(httpx.CommonResponseWriter{w})
			return
		}
		if httpres, ok := anyres.(httpx.IRespond); ok {
			httpres.Respond(w)
			return
		}
		w.Header().Set(httpx.HeaderContentType, httpx.ContentTypeJsonUtf8)
		json.NewEncoder(w).Encode(res)
	})
}
