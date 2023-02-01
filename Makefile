# Use nix (requires GNU make for SHELLFLAGS to work)
SHELL := nix-shell
.SHELLFLAGS := --run

.PHONY: test generate

test:
	go test -v -count=1 -cover -race ./...

test_ci:
	go test -v \
		-coverpkg=./... -coverprofile=coverage.out \
        -race ./...
	codecov

generate:
	go generate ./...

lint:
	golangci-lint run --timeout=3m --verbose
