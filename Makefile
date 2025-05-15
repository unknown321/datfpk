PRODUCT=datfpk
GOOS=linux
GOARCH=amd64
NAME=$(PRODUCT)-$(GOOS)-$(GOARCH)$(EXT)
EXT=
ifeq ($(GOOS),windows)
	override EXT=".exe"
endif

test:
	go test -v ./...

build:
	CGO_ENABLED=0 \
				GOOS=$(GOOS)\
				GOARCH=$(GOARCH) \
				go build -trimpath \
				-o $(NAME)

release: test
	$(MAKE) GOOS=windows build
	$(MAKE) GOOS=linux build
