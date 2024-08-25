// File:		basic_handler.go
// Created by:	Hoven
// Created on:	2024-08-09
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

import (
	"strings"
	
	"github.com/go-puzzles/puzzles/plog"
)

type HandleFunc func(ctx *Context) (Response, error)

func (f HandleFunc) Name() string {
	funcName := plog.GetFuncName(f)
	fs := strings.Split(funcName, ".")
	
	return fs[len(fs)-1]
}

func (f HandleFunc) Handle(ctx *Context) (Response, error) {
	return f(ctx)
}
