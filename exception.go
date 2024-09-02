// File:		exception.go
// Created by:	Hoven
// Created on:	2024-08-28
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

import (
	"fmt"
	"strings"
	
	"github.com/pkg/errors"
)

type Error interface {
	error
	
	Code() int
	Message() string
	Cause() error
	String() string
	Component() ErrComponent
	SetComponent(c ErrComponent) Error
	ResponseErrType() ResponseErrType
	SetResponseType(r ResponseErrType) Error
}

type ErrComponent string

const (
	ErrProuter  ErrComponent = "prouter"
	ErrRecovery ErrComponent = "recovery"
	ErrService  ErrComponent = "service"
	ErrRepo     ErrComponent = "repository"
	ErrLib      ErrComponent = "library"
)

type ResponseErrType string

const (
	BadRequest          ResponseErrType = "BadRequest"
	InternalServerError ResponseErrType = "InternalServerError"
	Forbidden           ResponseErrType = "Forbidden"
	NotFound            ResponseErrType = "NotFound"
	AlreadyExists       ResponseErrType = "AlreadyExists"
)

type prouterError struct {
	error
	code int
	// msg is the message that is displayed to the user
	msg          string
	cause        error
	component    ErrComponent
	responseType ResponseErrType
}

func NewErr(code int, err error, megs ...string) *prouterError {
	msg := strings.Join(megs, "\n")
	
	return &prouterError{code: code, msg: msg, cause: err}
}
func (e *prouterError) Error() string {
	var prefixes []string
	if e.component != "" {
		prefixes = append(prefixes, string(e.component))
	}
	
	if e.responseType != "" {
		prefixes = append(prefixes, string(e.responseType))
	}
	
	prefix := strings.Join(prefixes, ":")
	if prefix != "" {
		return fmt.Sprintf("[%s] Code: %d -> %s", prefix, e.code, e.Cause())
	}
	return fmt.Sprintf("Code: %d -> %s", e.code, e.Cause())
}

func (e *prouterError) Code() int {
	return e.code
}

func (e *prouterError) Message() string {
	msg := e.msg
	if msg == "" {
		msg = e.cause.Error()
	}
	return msg
}

func (e *prouterError) Cause() error {
	return e.cause
}

func (e *prouterError) String() string {
	return e.Error()
}

func (e *prouterError) Component() ErrComponent {
	return e.component
}

func (e *prouterError) SetComponent(comp ErrComponent) Error {
	e.component = comp
	return e
}

func (e *prouterError) ResponseErrType() ResponseErrType {
	return e.responseType
}

func (e *prouterError) SetResponseType(r ResponseErrType) Error {
	e.responseType = r
	return e
}

func ResourceAlreadyExists(code int, msg string) Error {
	return MsgError(code, msg).SetResponseType(AlreadyExists)
}

func ResourceNotFound(code int, msg string) Error {
	return MsgError(code, msg).SetResponseType(NotFound)
}

func MsgError(code int, msg string) Error {
	return NewErr(code, errors.New(msg))
}
