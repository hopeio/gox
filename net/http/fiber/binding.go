package fiber

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/hopeio/gox/strstruct"
	httpx "github.com/hopeio/gox/net/http"
)

func Bind(ctx *fiber.Ctx, obj any) error {
	return httpx.CommonBind(RequestSource{ctx}, obj)
}

type RequestSource struct {
	*fiber.Ctx
}

func (s RequestSource) Uri() strstruct.Getter {
	return strstruct.KVSource(s.AllParams())
}

func (s RequestSource) Query() strstruct.ValuesGetter {
	q := s.Queries()
	m := make(strstruct.KVsSource, len(q))
	for k, v := range q {
		m[k] = []string{v}
	}
	return m
}

func (s RequestSource) Header() strstruct.ValuesGetter {
	return httpx.HeaderSource(s.GetReqHeaders())
}

func (s RequestSource) Body() (context.Context, string, io.ReadCloser) {
	if s.Method() == http.MethodGet {
		return s.UserContext(), "", nil
	}
	return s.UserContext(), string(s.Request().Header.ContentType()), io.NopCloser(bytes.NewReader(s.Ctx.Body()))
}
