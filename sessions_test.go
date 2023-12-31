package main

import (
	"errors"

	session "github.com/fasthttp/session/v2"
	"github.com/valyala/fasthttp"
)

type dummySessionManager struct {
	getErr  bool
	saveErr bool
}

func (d *dummySessionManager) Get(*fasthttp.RequestCtx) (s *session.Store, err error) {
	s = session.NewStore()
	if d.getErr {
		err = errors.New("an error")
	}

	return
}

func (d *dummySessionManager) Save(*fasthttp.RequestCtx, *session.Store) (err error) {
	if d.saveErr {
		err = errors.New("an error")
	}

	return
}

func (d *dummySessionManager) SetProvider(session.Provider) error {
	return nil
}
