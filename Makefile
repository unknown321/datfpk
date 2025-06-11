PRODUCT=datfpk
GOOS=linux
GOARCH=amd64
NAME=$(PRODUCT)-$(GOOS)-$(GOARCH)$(EXT)
EXT=
ifeq ($(GOOS),windows)
	override EXT=".exe"
endif

IMAGE=golang:1.24.4-alpine
DOCKER=docker run -t --rm \
		-u $$(id -u):$$(id -g) \
		-v $$(pwd):$$(pwd) \
		-w $$(pwd) \
		-e GOCACHE=/tmp \
		-e CGO_ENABLED=0 \
		-e GOOS=$(GOOS)\
		-e GOARCH=$(GOARCH) \
		$(IMAGE)

test:
	$(DOCKER) go test -v ./...

build:
	$(DOCKER) go build -trimpath \
				-o $(NAME)

release: test
	$(MAKE) GOOS=windows build
	$(MAKE) GOOS=linux build
