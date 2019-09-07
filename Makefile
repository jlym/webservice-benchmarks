fmt:
	go fmt ./...

test:
	go test ./...

build: test build_server build_load_generator

build_server:
	go build -o bin/server github.com/jlym/webservice-benchmarks/cmd/httpserver

build_load_generator:
	go build -o bin/loadgenerator github.com/jlym/webservice-benchmarks/cmd/httpworker
	