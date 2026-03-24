package http

import (
	"context"
	"net/http"

	"github.com/hopeio/gox/errors"
	"github.com/hopeio/gox/types"
)

type Service[REQ, RESP any] func(ctx ReqResp, req REQ) (RESP, *ErrResp)

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

func HandlerWrap[REQ, RESP any](service Service[*REQ, *RESP]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := new(REQ)
		err := Bind(r, req)
		if err != nil {
			ServeError(w, r, errors.InvalidArgument.Msg(err.Error()))
			return
		}
		res, err := service(ReqResp{r, w}, req)
		if err != nil {
			ServeError(w, r, err)
			return
		}
		switch httpres := any(res).(type) {
		case http.Handler:
			httpres.ServeHTTP(w, r)
			return
		case Responder:
			httpres.Respond(r.Context(), w)
			return
		}
		ServeSuccess(w, r, res)
	})
}
func HandlerWrapGRPC[REQ, RESP any](method types.GrpcService[*REQ, *RESP]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		req := new(REQ)
		err := Bind(r, req)
		if err != nil {
			ServeError(w, r, errors.InvalidArgument.Wrap(err))
			return
		}
		res, err := method(WrapContext(ReqResp{r, w}), req)
		if err != nil {
			ErrRespFrom(err).ServeHTTP(w, r)
			return
		}
		switch httpres := any(res).(type) {
		case http.Handler:
			httpres.ServeHTTP(w, r)
			return
		case Responder:
			httpres.Respond(r.Context(), w)
			return
		}
		ServeSuccess(w, r, res)
	})
}
