CMD=go build
SRC=./cmd
DST=./build
NAME=gosnake

default: build

fmt:
	go fmt ./...
	go mod tidy

build_linux: fmt
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(CMD) -o $(DST)/$(NAME)_linux $(SRC)  

build_darwin: fmt
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(CMD) -o $(DST)/$(NAME)_darwin $(SRC)

build_windows: fmt
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(CMD) -o $(DST)/$(NAME)_windows.exe $(SRC)

build_all: build_linux build_darwin build_windows

build: fmt
	$(CMD) -o $(DST)/$(NAME) $(SRC)
.PHONY: build
	

run: fmt
	go run ./cmd/main.go

clean:
	rm -rf $(DST)/*
