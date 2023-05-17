peer-demedia: $(shell find .. -name "*.go")
	CC=$$(which musl-gcc) go build -ldflags='-s -w -linkmode external -extldflags "-static"' -o ./peer-demedia
