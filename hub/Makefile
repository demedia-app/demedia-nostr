demedia-hub: $(shell find .. -name "*.go")
	CC=$$(which musl-gcc) CGO_ENABLED=0 go build -tags netgo -ldflags='-s -w -linkmode external -extldflags "-static"' -o ./demedia-hub
