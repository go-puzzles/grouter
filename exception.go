// File:		exception.go
// Created by:	Hoven
// Created on:	2024-08-28
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

import "fmt"

type prouterError struct {
	code int
	msg  string
}

func (e *prouterError) Error() string {
	return fmt.Sprintf("prouter error. statusCode: %s, msg: %s", e.code, e.msg)
}

func (e *prouterError) Code() int {
	return e.code
}

func (e *prouterError) Message() string {
	return e.msg
}

func NewError(statusCode int, msg string) *prouterError {
	return &prouterError{code: statusCode, msg: msg}
}
