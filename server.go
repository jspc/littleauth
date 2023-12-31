package main

import (
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

const (
	// isLoggedIn is the key used in the session manager to signify whether
	// the user is, uhhhh, logged in
	isLoggedIn = "is-logged-in"
)

type Server struct {
	config Config

	*router.Router
}

func New(c Config) (s Server, err error) {
	s.config = c
	s.Router = router.New()

	s.GET("/", s.RenderForm)

	api := s.Group("/api")

	v1 := api.Group("/v1")
	v1.GET("/auth", s.Auth)
	v1.POST("/login", s.Login)

	return
}

// Auth determines whether a request is allowed access to the specified downstream
// and either:
//
//  1. Returns a 200, signifying a request is valid; or
//  2. Returns a 303 to the login form
//
// We use a 303 to ensure that the form is always requested as a GET
func (s Server) Auth(ctx *fasthttp.RequestCtx) {
	vhost := s.config.MatchVHost(ctx.Request.Header.Peek("X-Forwarded-Host"))

	store, err := vhost.sm.Get(ctx)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)

		return
	}

	if b := store.Get(isLoggedIn); b != nil {
		ctx.SetStatusCode(fasthttp.StatusOK)

		return
	}

	ctx.Redirect("https://auth.beasts.jspc.pw", fasthttp.StatusSeeOther)
}

// Login is a form handler, which receives a username and password from the login form.
//
// It looks these up against the htpasswd specified in the app config. Where the
// credentials are correct, we redirect to the original URL. Where they don't, we
// return a very unceremonious 403 message
func (s Server) Login(ctx *fasthttp.RequestCtx) {
	username := string(ctx.FormValue("username"))
	password := string(ctx.FormValue("password"))

	vhost := s.config.MatchVHost(ctx.Request.Header.Peek("X-Forwarded-Host"))

	if vhost.passwd.Match(username, password) {
		store, err := vhost.sm.Get(ctx)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}

		store.Set(isLoggedIn, true)
		err = vhost.sm.Save(ctx, store)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)

			return
		}

		ctx.Redirect(vhost.Redirect, fasthttp.StatusSeeOther)

		return
	}

	ctx.Error("incorrect username/password combination", fasthttp.StatusForbidden)
}

// RenderForm shows the login form as provided by the sysadmin
func (s Server) RenderForm(ctx *fasthttp.RequestCtx) {
	vhost := s.config.MatchVHost(ctx.Request.Header.Peek("X-Forwarded-Host"))

	ctx.SetContentType("text/html; charset=utf8")

	err := vhost.templates.Execute(ctx, "login")
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}
