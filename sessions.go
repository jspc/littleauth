package main

import (
	session "github.com/fasthttp/session/v2"
	"github.com/valyala/fasthttp"
)

type SessionManager interface {
	Get(*fasthttp.RequestCtx) (*session.Store, error)
	Save(*fasthttp.RequestCtx, *session.Store) error
	SetProvider(session.Provider) error
}
