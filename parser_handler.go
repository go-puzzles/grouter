// File:		handler.go
// Created by:	Hoven
// Created on:	2024-08-09
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-puzzles/plog"
)

type bodyParseHandlerFn[RequestT any, ResponseT any] func(context.Context, *RequestT) (*ResponseT, error)

func BodyParserHandleFunc[RequestT any, ResponseT any](fn func(context.Context, *RequestT) (*ResponseT, error)) bodyParseHandlerFn[RequestT, ResponseT] {
	return bodyParseHandlerFn[RequestT, ResponseT](fn)
}

func (h bodyParseHandlerFn[RequestT, ResponseT]) Name() string {
	funcName := plog.GetFuncName(h)
	fs := strings.Split(funcName, ".")

	return fs[len(fs)-1]
}

func (h bodyParseHandlerFn[RequestT, ResponseT]) Handle(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (resp Response, err error) {
	ret := &Ret{}
	requestPtr := new(RequestT)
	// requestPtr := reflect.New(tml).Interface()

	binder := binding.Default(r.Method, contentType(r))
	if err = binder.Bind(r, requestPtr); err != nil {
		ret.SetMessage("parse request data failed")
		ret.SetCode(http.StatusBadRequest)
		return nil, err
	}

	if len(vars) > 0 {
		m := make(map[string][]string)
		for k, v := range vars {
			m[k] = []string{v}
		}
		if err = binding.Uri.BindUri(m, requestPtr); err != nil {
			ret.SetMessage("parse request uri data failed")
			ret.SetCode(http.StatusBadRequest)
			return nil, err
		}
	}

	if len(r.Header) > 0 {
		if err = binding.Header.Bind(r, requestPtr); err != nil {
			ret.SetMessage("parse request header data failed")
			ret.SetCode(http.StatusBadRequest)
			return nil, err
		}
	}

	if len(r.URL.Query()) > 0 {
		if err = binding.Query.Bind(r, requestPtr); err != nil {
			ret.SetMessage("parse request query data failed")
			ret.SetCode(http.StatusBadRequest)
			return nil, err
		}
	}

	handleResp, err := h(ctx, requestPtr)
	if err != nil {
		ret.SetMessage(err.Error())
		ret.SetCode(http.StatusInternalServerError)
		return nil, err
	}

	ret.SetData(handleResp)
	ret.SetCode(http.StatusOK)

	return ret, err
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
