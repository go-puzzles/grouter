// File:		main.go
// Created by:	Hoven
// Created on:	2024-08-20
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package main

import (
	"fmt"
	"net/http"

	"github.com/go-puzzles/plog"
	"github.com/go-puzzles/predis"
	"github.com/go-puzzles/prouter"
	sessionstore "github.com/go-puzzles/prouter/session-store"
	"github.com/pkg/errors"
)

func helloHandler(ctx *prouter.Context) (prouter.Response, error) {
	data, err := ctx.Session.Get("name")
	if errors.Is(err, prouter.SessionKeyNotExists) {
		ctx.Session.Set("name", "super")
		return prouter.SuccessResponse("hello world"), nil
	} else if err != nil {
		return nil, errors.Wrap(err, "get")
	}

	return prouter.SuccessResponse(fmt.Sprintf("hello world : %v", data.(string))), nil
}

func main() {
	redisConf := &predis.RedisConf{}
	redisConf.SetDefault()

	pool := redisConf.DialRedisPool()
	redisStore := sessionstore.NewRedisStore(pool, "sesionprefix")

	router := prouter.NewProuter()
	router.UseMiddleware(prouter.NewSessionMiddleware("testsession", redisStore))

	router.GET("/hello", prouter.HandleFunc(helloHandler))

	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
	}
	plog.PanicError(srv.ListenAndServe())

}
