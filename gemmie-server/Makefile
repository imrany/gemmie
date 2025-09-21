all: build run

delete: 
	rm -rf bin

build:
	$(MAKE) delete
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/gemmie-server-linux main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o bin/gemmie-server-windows.exe main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o bin/gemmie-server-darwin main.go   

run:
	./bin/gemmie-server-linux

ensure-compile-daemon:
	@which go > /dev/null || (echo "Error: Go is not installed or not in PATH" && exit 1)
	@which CompileDaemon > /dev/null || (echo "Installing CompileDaemon..." && go install github.com/githubnemo/CompileDaemon@latest)

dev:
	CompileDaemon -build="go build -o ./bin/gemmie-server main.go" -command="./bin/gemmie-server"