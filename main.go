package main

import (
	"os"

	"github.com/valyala/fasthttp"
)

var (
	cfgFile = envOrDefault("CONFIG", "testdata/sample-config.toml")
)

func main() {
	s, err := prepareServer()
	if err != nil {
		panic(err)
	}

	panic(fasthttp.ListenAndServe(":8080", s.Handler))
}

func prepareServer() (s *Server, err error) {
	c, err := ReadConfig(cfgFile)
	if err != nil {
		return
	}

	return New(c)
}

func envOrDefault(v, d string) string {
	s, ok := os.LookupEnv(v)
	if ok {
		return s
	}

	return d
}
