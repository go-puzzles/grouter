// File:		main.go
// Created by:	Hoven
// Created on:	2024-08-30
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package main

import (
	"fmt"
	"net/http"

	"github.com/go-puzzles/prouter"
	"github.com/go-puzzles/puzzles/plog"
)

func Middleware1(ctx *prouter.Context) (prouter.Response, error) {
	fmt.Println("middleware 1")
	return nil, nil
}

func main() {
	prouter.SetMode(prouter.DebugMode)
	router := prouter.NewProuter()
	router.Use(Middleware1)

	router.GET("/test", prouter.HandleFunc(func(ctx *prouter.Context) (prouter.Response, error) {
		fmt.Println("test router")
		return nil, nil
	}))

	grp := router.Group("/grp")
	grp.Use(Middleware1)
	grp.GET("test", prouter.HandleFunc(func(ctx *prouter.Context) (prouter.Response, error) {
		fmt.Println("test grp router")
		return nil, nil
	}))

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	plog.PanicError(srv.ListenAndServe())
}
