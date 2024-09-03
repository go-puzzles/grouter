// File:		handler.go
// Created by:	Hoven
// Created on:	2024-08-09
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-puzzles/puzzles/plog"
	"github.com/pkg/errors"
)

type handlerFunc interface {
	Name() string
	Handle(ctx *Context) (Response, error)
}

type wrapHandler struct {
	name    string
	handler HandleFunc
}

func (h *wrapHandler) Name() string {
	return h.name
}

func (h *wrapHandler) Handle(ctx *Context) (Response, error) {
	return h.handler.Handle(ctx)
}

type HandleFunc func(ctx *Context) (Response, error)

func (f HandleFunc) WrapHandler(handler handlerFunc) handlerFunc {
	return HandleFunc(func(ctx *Context) (Response, error) {
		resp, err := f(ctx)
		if err != nil {
			return resp, err
		}

		return handler.Handle(ctx)
	})
}

func (f HandleFunc) Name() string {
	funcName := plog.GetFuncName(f)
	fs := strings.Split(funcName, ".")

	return fs[len(fs)-1]
}

func (f HandleFunc) Handle(ctx *Context) (Response, error) {
	return f(ctx)
}

type bodyParseHandlerFn[RequestT any, ResponseT any] func(*Context, *RequestT) (*ResponseT, error)

func BodyParser[RequestT any, ResponseT any](fn func(*Context, *RequestT) (*ResponseT, error)) HandleFunc {
	return bodyParseHandlerFn[RequestT, ResponseT](fn).Handle
}

func (h bodyParseHandlerFn[RequestT, ResponseT]) Name() string {
	funcName := plog.GetFuncName(h)
	fs := strings.Split(funcName, ".")

	return fs[len(fs)-1]
}

func (h bodyParseHandlerFn[RequestT, ResponseT]) Handle(ctx *Context) (resp Response, err error) {
	ret := NewResponseTmpl()
	requestPtr := new(RequestT)
	r := ctx.Request

	var errMsg string
	func() {
		ct := contentType(r)
		binder := binding.Default(r.Method, ct)
		if err = binder.Bind(r, requestPtr); err != nil {
			errMsg = "parse request data failed"
		}
		if len(ctx.vars) > 0 {
			m := make(map[string][]string)
			for k, v := range ctx.vars {
				m[k] = []string{v}
			}
			if err = binding.Uri.BindUri(m, requestPtr); err != nil {
				errMsg = "parse request data failed"
				return
			}
		}

		if len(r.Header) > 0 {
			if err = binding.Header.Bind(r, requestPtr); err != nil {
				errMsg = "parse request header data failed"
				return
			}
		}

		if len(r.URL.Query()) > 0 {
			if err = binding.Query.Bind(r, requestPtr); err != nil {
				errMsg = "parse request query data failed"
				return
			}
		}
	}()

	if errMsg != "" {
		return nil, NewErr(http.StatusBadRequest, err, errMsg).
			SetComponent(ErrProuter).
			SetResponseType(BadRequest)
	}

	handleResp, err := h(ctx, requestPtr)
	if err != nil {
		routerErr := new(prouterError)
		if errors.As(err, &routerErr) {
			return nil, routerErr
		}
		return nil, NewErr(http.StatusBadRequest, err).
			SetComponent(ErrService).
			SetResponseType(BadRequest)
	}

	ret.SetData(handleResp)
	ret.SetCode(http.StatusOK)

	return ret, nil
}

func contentType(r *http.Request) string {
	return filterFlags(requestHeader(r, "Content-Type"))
}

func requestHeader(req *http.Request, key string) string {
	return req.Header.Get(key)
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
