package main

import (
	"html/template"
	"path/filepath"
	"time"

	"github.com/fasthttp/router"
	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/tg123/go-htpasswd"
	"github.com/valyala/fasthttp"
)

const (
	// isLoggedIn is the key used in the session manager to signify whether
	// the user is, uhhhh, logged in
	isLoggedIn = "is-logged-in"
)

type Server struct {
	templates    *template.Template
	sm           *session.Session
	passwd       *htpasswd.File
	authRedirect string

	*router.Router
}

func New(passwordPath, templatesPath, authRedirect string) (s Server, err error) {
	s.authRedirect = authRedirect
	s.Router = router.New()

	s.passwd, err = htpasswd.New(passwordPath, htpasswd.DefaultSystems, nil)
	if err != nil {
		return
	}

	s.templates, err = template.New("login").ParseGlob(filepath.Join(templatesPath, "*.html.tmpl"))
	if err != nil {
		return
	}

	err = s.setSessionManager()
	if err != nil {
		return
	}

	s.GET("/", s.RenderForm)

	api := s.Group("/api")

	v1 := api.Group("/v1")
	v1.GET("/auth", s.Auth)
	v1.POST("/login", s.Login)

	return
}

func (s *Server) setSessionManager() (err error) {
	cfg := session.NewDefaultConfig()
	cfg.CookieName = "littleauth"
	cfg.Domain = "jspc.pw"
	cfg.Expiration = time.Second * 604800
	cfg.Secure = true

	cfg.EncodeFunc = session.MSGPEncode
	cfg.DecodeFunc = session.MSGPDecode

	s.sm = session.New(cfg)

	provider, err := memory.New(memory.Config{})
	if err != nil {
		return
	}

	return s.sm.SetProvider(provider)
}

// Auth determines whether a request is allowed access to the specified downstream
// and either:
//
//  1. Returns a 200, signifying a request is valid; or
//  2. Returns a 303 to the login form
//
// We use a 303 to ensure that the form is always requested as a GET
func (s Server) Auth(ctx *fasthttp.RequestCtx) {
	store, err := s.sm.Get(ctx)
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

	if s.passwd.Match(username, password) {
		store, err := s.sm.Get(ctx)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		}

		store.Set(isLoggedIn, true)
		err = s.sm.Save(ctx, store)
		if err != nil {
			ctx.Error(err.Error(), fasthttp.StatusInternalServerError)

			return
		}

		ctx.Redirect(s.authRedirect, fasthttp.StatusSeeOther)

		return
	}

	ctx.Error("incorrect username/password combination", fasthttp.StatusForbidden)
}

// RenderForm shows the login form as provided by the sysadmin
func (s Server) RenderForm(ctx *fasthttp.RequestCtx) {
	ctx.SetContentType("text/html; charset=utf8")

	err := s.templates.Execute(ctx, "login")
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
	}
}
