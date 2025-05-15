test:
	go test -v ./...

build:
	CGO_ENABLED=0 go build -trimpath .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath .
