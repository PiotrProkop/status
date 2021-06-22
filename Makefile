.PHONY: build clean

build:
	go build -o _output/status-server ./cmd/status-server 

build-docker:
	docker build -f build/Dockerfile -t status-server:latest .

clean:
	rm -f _output/*
lint:
	golangci-lint run

test:
	go test -v ./...

