fmt:
	go fmt ./...

test:
	go test ./...

build: test build_server build_load_generator build_monitor

build_load_generator:
	go build -o bin/load_generator github.com/jlym/webservice-benchmarks/cmd/load_generator

build_monitor:
	go build -o bin/monitor github.com/jlym/webservice-benchmarks/cmd/monitor

build_server:
	go build -o bin/server github.com/jlym/webservice-benchmarks/cmd/httpserver

