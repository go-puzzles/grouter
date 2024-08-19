// File:		context.go
// Created by:	Hoven
// Created on:	2024-08-19
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

import (
	"context"
	"net/http"
	"time"
)

type ContextKeyType int

const (
	ContextRequestKey ContextKeyType = iota
	ContextKey
)

type Context struct {
	context.Context
	Request    *http.Request
	Writer     http.ResponseWriter
	Path       string
	ClientIp   string
	Method     string
	StatusCode int

	startTime time.Time
}

func (c *Context) Ctx() context.Context {
	return c.Context
}
