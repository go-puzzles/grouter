// File:		redis_store.go
// Created by:	Hoven
// Created on:	2024-08-20
//
// This file is part of the Example Project.
//
// (c) 2024 Example Corp. All rights reserved.

package sessionstore

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/go-puzzles/predis"
	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

// SessionSerializer provides an interface for serialize/deserialize a session
type SessionSerializer interface {
	Serialize(s *sessions.Session) ([]byte, error)
	Deserialize(b []byte, s *sessions.Session) error
}

type GobSerializer struct{}

func (gs GobSerializer) Serialize(s *sessions.Session) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(s.Values)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (gs GobSerializer) Deserialize(d []byte, s *sessions.Session) error {
	dec := gob.NewDecoder(bytes.NewBuffer(d))
	return dec.Decode(&s.Values)
}

type RedisStore struct {
	serializer SessionSerializer
	Options    *sessions.Options

	client *predis.RedisClient
	prefix string
}

func NewRedisStore(pool *redis.Pool, prefix string) *RedisStore {
	return &RedisStore{
		client:     predis.NewRedisClient(pool),
		serializer: &GobSerializer{},
		prefix:     prefix,
		Options: &sessions.Options{
			Path:     "/",
			MaxAge:   86400 * 30,
			SameSite: http.SameSiteNoneMode,
			Secure:   true,
		},
	}
}

func (s *RedisStore) Key(k string) string {
	return fmt.Sprintf("%s:%s", s.prefix, k)
}

// Get get a session from redis
func (s *RedisStore) Get(r *http.Request, name string) (session *sessions.Session, err error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New create a new session
func (s *RedisStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := s.Options
	session.Options = opts
	session.IsNew = true

	c, err := r.Cookie(name)
	if err != nil {
		return session, nil
	}
	session.ID = c.Value

	err = s.load(session)
	if err == nil {
		session.IsNew = false
	} else if errors.Is(err, redis.ErrNil) {
		err = nil // no data stored
	}
	return session, err
}

// Save session to redis
func (s *RedisStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	// Delete if max-age is <= 0
	if session.Options.MaxAge <= 0 {
		if err := s.delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		id, err := s.generateRandomKey()
		if err != nil {
			return errors.New("redisstore: failed to generate session id")
		}
		session.ID = id
	}
	if err := s.save(session); err != nil {
		return err
	}

	http.SetCookie(w, sessions.NewCookie(session.Name(), session.ID, session.Options))
	return nil
}

// save writes session in Redis
func (s *RedisStore) save(session *sessions.Session) error {
	b, err := s.serializer.Serialize(session)
	if err != nil {
		return err
	}

	return s.client.SetWithTTL(s.Key(session.ID), b, time.Duration(session.Options.MaxAge)*time.Second)
}

func (s *RedisStore) load(session *sessions.Session) error {
	var data []byte
	err := s.client.Get(s.Key(session.ID), &data)
	if err != nil {
		return errors.Wrap(err, "getRedis")
	}

	if err := s.serializer.Deserialize(data, session); err != nil {
		return errors.Wrap(err, "deserialize")
	}
	return nil
}

// delete deletes session in Redis
func (s *RedisStore) delete(session *sessions.Session) error {
	return s.client.Delete(s.Key(session.ID))
}

// generateRandomKey returns a new random key
func (s *RedisStore) generateRandomKey() (string, error) {
	return uuid.NewString(), nil
	// k := make([]byte, 64)
	// if _, err := io.ReadFull(rand.Reader, k); err != nil {
	// 	return "", err
	// }
	// return strings.TrimRight(base32.StdEncoding.EncodeToString(k), "="), nil
}
