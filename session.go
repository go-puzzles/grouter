// File:		session.go
// Created by:	Hoven
// Created on:	2024-08-20
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package prouter

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
)

const (
	sessionKey              = "prouter:session"
	defaultSessionSecretKey = "prouter:session-secret-key"
)

var (
	SessionNotInitialized = fmt.Errorf("Session not initialized")
	SessionKeyNotExists   = fmt.Errorf("Session key not exist")
)

type session struct {
	key     string
	session *sessions.Session

	r *http.Request
	w http.ResponseWriter
}

func (s *session) Save() error {
	if s == nil {
		return SessionNotInitialized
	}
	return s.session.Save(s.r, s.w)
}

func (s *session) ID() string {
	if s == nil {
		return ""
	}
	return s.session.ID
}

func (s *session) Get(key string) (interface{}, error) {
	if s == nil {
		return nil, SessionNotInitialized
	}

	val, exists := s.session.Values[key]
	if !exists {
		return nil, SessionKeyNotExists
	}
	return val, nil
}

func (s *session) Set(key string, value interface{}) error {
	if s == nil {
		return SessionNotInitialized
	}

	s.session.Values[key] = value
	return nil
}

type SessionMiddleware struct {
	key   string
	store sessions.Store
}

func NewSessionMiddleware(key string, stores ...sessions.Store) *SessionMiddleware {
	var store sessions.Store
	if len(stores) == 0 {
		store = sessions.NewCookieStore([]byte(defaultSessionSecretKey))
	} else {
		store = stores[0]
	}

	return &SessionMiddleware{
		key:   key,
		store: store,
	}
}

func (m *SessionMiddleware) WrapHandler(handler handlerFunc) handlerFunc {
	return HandleFunc(func(ctx *Context) (resp Response, err error) {
		s, err := m.store.Get(ctx.Request, m.key)
		if err != nil {
			return nil, err
		}

		sess := &session{
			key:     m.key,
			session: s,
			r:       ctx.Request,
			w:       ctx.Writer,
		}
		ctx.Session = sess
		defer func() {
			if newErr := ctx.Session.Save(); newErr != nil {
				err = errors.Join(err, newErr)
				return
			}
		}()

		resp, err = handler.Handle(ctx)
		return
	})
}
