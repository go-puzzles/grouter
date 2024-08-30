// File:		middleware.go
// Created by:	Hoven
// Created on:	2024-08-30
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

type Middleware interface {
	WrapHandler(handler handlerFunc) handlerFunc
}
