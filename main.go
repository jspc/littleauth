package main

import (
	"os"

	"github.com/valyala/fasthttp"
)

var (
	passwdFile       = envOrDefault("PASSWD_FILE", "testdata/passwd")
	formsDir         = envOrDefault("FORMS_DIR", "testdata/form")
	postAuthRedirect = envOrDefault("POST_AUTH_REDIR", "https://ipv4.rides.beasts.jspc.pw")
)

func main() {
	s, err := New(passwdFile, formsDir, postAuthRedirect)
	if err != nil {
		panic(err)
	}

	panic(fasthttp.ListenAndServe(":8080", s.Handler))
}

func envOrDefault(v, d string) string {
	s, ok := os.LookupEnv(v)
	if ok {
		return s
	}

	return d
}
