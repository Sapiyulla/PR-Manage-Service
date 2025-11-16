run:
	go run cmd/server/main.go
	

build:
	go build -ldflags="-s -w" cmd/server/main.go

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/app cmd/server/main.go

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/app cmd/server/main.go
	
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/app cmd/server/main.go
