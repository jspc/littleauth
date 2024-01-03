package main

import (
	"net/http"
	_ "net/http/pprof" //#nosec: G108
	"os"
)

func init() {
	go func() {
		if os.Getenv("COLLECT_METRICS") == "yes" {
			//#nosec: G114
			print(http.ListenAndServe("localhost:6060", nil), "\n")
		}
	}()
}
