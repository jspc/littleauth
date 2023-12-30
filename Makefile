go_build_cmd ?= CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath

littleauth: *.go go.*
	$(go_build_cmd) -o $@
	-upx $@
