// File:		basic_handler.go
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

	"github.com/go-puzzles/plog"
)

type HandleFunc func(ctx *Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (Response, error)

func (f HandleFunc) Name() string {
	funcName := plog.GetFuncName(f)
	fs := strings.Split(funcName, ".")

	return fs[len(fs)-1]
}

func (f HandleFunc) Handle(ctx *Context, w http.ResponseWriter, r *http.Request, vars map[string]string) (Response, error) {
	return f(ctx, w, r, vars)
}
