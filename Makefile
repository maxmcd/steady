# Use nix (requires GNU make for SHELLFLAGS to work)
SHELL := nix-shell
.SHELLFLAGS := --run

.PHONY: test generate

test:
	go test -v -count=1 -cover ./...

test_ci:
	STEADY_SUITE_RUN_COUNT=10 go test -v \
		-coverpkg=./... -coverprofile=coverage.out \
        -race ./...
	codecov

generate:
	go generate -x ./...

lint:
	golangci-lint run --timeout=3m
