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

type wrapKey struct{}

var wrapContextKey = wrapKey{}

func WrapContext(v any) context.Context {
	return context.WithValue(context.Background(), wrapContextKey, v)
}

func UnWrapContext(ctx context.Context) any {
	return ctx.Value(wrapContextKey)
}

type ReqResp struct {
	*http.Request
	http.ResponseWriter
}

func HandlerWrap[REQ, RES any](service Service[*REQ, *RES]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := new(REQ)
		err := binding.Bind(r, req)
		if err != nil {
			httpx.RespErrCodeMsg(ctx, w, errors.InvalidArgument, err.Error())
			return
		}
		res, errRep := service(ReqResp{r, w}, req)
		if err != nil {
			errRep.Respond(ctx, w)
			return
		}
		switch httpres := any(res).(type) {
		case http.Handler:
			httpres.ServeHTTP(w, r)
			return
		case httpx.CommonResponder:
			httpres.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: w})
			return
		case httpx.Responder:
			httpres.Respond(ctx, w)
			return
		}
		json.NewEncoder(w).Encode(res)
	})
}
func HandlerWrapGRPC[REQ, RES any](method types.GrpcService[*REQ, *RES]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := new(REQ)
		err := binding.Bind(r, req)
		if err != nil {
			httpx.RespSuccessData(ctx, w, errors.InvalidArgument.Wrap(err))
			return
		}
		res, err := method(WrapContext(ReqResp{r, w}), req)
		if err != nil {
			httpx.ErrRespFrom(err).Respond(ctx, w)
			return
		}
		switch httpres := any(res).(type) {
		case http.Handler:
			httpres.ServeHTTP(w, r)
			return
		case httpx.CommonResponder:
			httpres.CommonRespond(ctx, httpx.ResponseWriterWrapper{ResponseWriter: w})
			return
		case httpx.Responder:
			httpres.Respond(ctx, w)
			return
		}
		w.Header().Set(httpx.HeaderContentType, httpx.ContentTypeJsonUtf8)
		json.NewEncoder(w).Encode(res)
	})
}
