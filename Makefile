go_build_cmd ?= CGO_ENABLED=1 go build -ldflags="-s -w" -trimpath -race -pgo merged.pprof

littleauth: *.go go.*
	$(go_build_cmd) -o $@
	-upx $@
