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
	"reflect"

	"github.com/gin-gonic/gin/binding"
)

type BodyParseHandlerFn[RequestT any, ResponseT any] func(context.Context, *RequestT) (*ResponseT, error)

func NewBodyParseHandler[RequestT, ResponseT any](fn BodyParseHandlerFn[RequestT, ResponseT]) HandleFunc {
	tml := reflect.TypeOf((*RequestT)(nil))
	if tml.Kind() != reflect.Pointer {
		panic("RequestT must be a pointer")
	}

	tml = tml.Elem()
	if tml.Kind() != reflect.Struct {
		panic("RequestT must be a struct")
	}

	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (resp Response, err error) {
		ret := &Ret{}
		requestPtr := reflect.New(tml).Interface()

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

		handleResp, err := fn(ctx, requestPtr.(*RequestT))
		if err != nil {
			ret.SetMessage(err.Error())
			ret.SetCode(http.StatusInternalServerError)
			return nil, err
		}

		ret.SetData(handleResp)
		ret.SetCode(http.StatusOK)

		return ret, err
	}
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
