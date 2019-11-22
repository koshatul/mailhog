DOCKER_REPO ?= koshatul/mailhog

GO_MATRIX_OS := darwin linux
#freebsd linux netbsd openbsd windows
GO_MATRIX_ARCH := 386 amd64

GENERATED_FILES += ui/assets/assets.go

DOCKER_BUILD_REQ += $(shell find ./ -name '*.go')
DOCKER_BUILD_REQ += $(GENERATED_FILES)

-include .makefiles/Makefile
-include .makefiles/pkg/go/v1/Makefile
-include .makefiles/pkg/docker/v1/Makefile

.makefiles/%:
	@curl -sfL https://makefiles.dev/v1 | bash /dev/stdin "$@"

######################
## UI Assets
######################

GO_BINDATA := artifacts/go-bindata/bin/go-bindata
$(GO_BINDATA):
	GO111MODULE=off GOBIN="$(MF_PROJECT_ROOT)/$(@D)" go get github.com/jteeuwen/go-bindata/...

_ASSETS := $(shell find ui/assets/ -type f)

ui/assets/assets.go: $(_ASSETS) | $(GO_BINDATA)
	-rm "$(@)"
	cd ui && ../$(GO_BINDATA) -o assets/assets.go -pkg assets assets/...
