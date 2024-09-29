go_build_cmd ?= CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -race -pgo merged.pprof
#go_build_cmd ?= tinygo build -cpuprofile merged.pprof -no-debug

littleauth: *.go go.* merged.pprof
	$(go_build_cmd) -o $@
	-upx $@

merged.pprof:
	go tool pprof -proto -output merged.pprof profiles/*.pprof
