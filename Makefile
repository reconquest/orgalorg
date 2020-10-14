build:
	go build -mod=mod -ldflags -X=main.version=$$(git describe --tags --abbrev=6)

test:
	./run_tests -vvvv
