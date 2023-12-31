package main

import (
	"os"

	"github.com/valyala/fasthttp"
)

var (
	cfgFile = envOrDefault("CONFIG", "testdata/sample-config.toml")
)

func main() {
	c, err := ReadConfig(cfgFile)
	if err != nil {
		panic(err)
	}

	s, err := New(c)
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
