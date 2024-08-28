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
	"html/template"
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
	router *Prouter
	vars   map[string]string

	Request    *http.Request
	Writer     http.ResponseWriter
	Path       string
	ClientIp   string
	Method     string
	StatusCode int

	Session *session

	startTime time.Time
}

func (c *Context) Ctx() context.Context {
	return c.Context
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
