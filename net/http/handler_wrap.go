package http

import (
	"context"
	"net/http"

	"github.com/hopeio/gox/errors"
	"github.com/hopeio/gox/types"
)

type Service[REQ, RES any] func(ctx ReqResp, req REQ) (RES, *ErrResp)

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
		err := Bind(r, req)
		if err != nil {
			RespondError(ctx, w, errors.InvalidArgument.Msg(err.Error()))
			return
		}
		res, errResp := service(ReqResp{r, w}, req)
		if err != nil {
			RespondError(ctx, w, errResp)
			return
		}
		switch httpres := any(res).(type) {
		case http.Handler:
			httpres.ServeHTTP(w, r)
			return
		case CommonResponder:
			httpres.CommonRespond(ctx, ResponseWriterWrapper{ResponseWriter: w})
			return
		case Responder:
			httpres.Respond(ctx, w)
			return
		}
		RespondSuccess(ctx, w, res)
	})
}
func HandlerWrapGRPC[REQ, RES any](method types.GrpcService[*REQ, *RES]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := new(REQ)
		err := Bind(r, req)
		if err != nil {
			RespondError(ctx, w, errors.InvalidArgument.Wrap(err))
			return
		}
		res, err := method(WrapContext(ReqResp{r, w}), req)
		if err != nil {
			ErrRespFrom(err).Respond(ctx, w)
			return
		}
		switch httpres := any(res).(type) {
		case http.Handler:
			httpres.ServeHTTP(w, r)
			return
		case CommonResponder:
			httpres.CommonRespond(ctx, ResponseWriterWrapper{ResponseWriter: w})
			return
		case Responder:
			httpres.Respond(ctx, w)
			return
		}
		RespondSuccess(ctx, w, res)
	})
}
