run:
	go run cmd/server/main.go
	

build:
	go build -ldflags="-s -w" cmd/server/main.go
	.\main.exe

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" cmd/server/main.go
	main

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" cmd/server/main.go
	main
	
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" cmd/server/main.go
	main
