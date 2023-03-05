

build_runner:
	docker build -t runner -f ./runner/Dockerfile  ../../

test: build_runner
	go test -v -race -count=1  .
