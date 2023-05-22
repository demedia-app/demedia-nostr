demedia-peer: $(shell find .. -name "*.go")
	CC=$$(which musl-gcc) go build -tags netgo -ldflags='-s -w -linkmode external -extldflags "-static"' -o ./demedia-peer
