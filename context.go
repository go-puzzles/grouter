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
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"
	
	"github.com/go-puzzles/puzzles/plog"
)

type ContextKeyType int

const (
	ContextRequestKey ContextKeyType = iota
	ContextKey
)

type Context struct {
	context.Context
	router *Prouter
	vars   map[string]string
	
	Request    *http.Request
	Writer     http.ResponseWriter
	Path       string
	ClientIp   string
	Method     string
	StatusCode int
	
	session *Session
	
	startTime time.Time
}

func (c *Context) Ctx() context.Context {
	return c.Context
}

func (c *Context) WithValue(key, val any) {
	c.Context = context.WithValue(c.Context, key, val)
}

func (c *Context) Session() *Session {
	if c.session == nil {
		plog.PanicError(fmt.Errorf("Session not initialized"))
	}
	return c.session
}

func (c *Context) ExecuteTemplateFS(fs embed.FS, resource string, data any) (Response, error) {
	tmpl, err := template.ParseFS(fs, resource)
	if err != nil {
		return nil, err
	}
	
	err = tmpl.Execute(c.Writer, data)
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

func (c *Context) Redirect(code int, location string) (Response, error) {
	http.Redirect(c.Writer, c.Request, location, code)
	return nil, nil
}
