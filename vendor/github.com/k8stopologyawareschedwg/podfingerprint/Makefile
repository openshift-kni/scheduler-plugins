all: test-unit


test-unit:
	go test ./...

pfp:
	go build -o _out/pfp -v ./tools/pfp/...

clean:
	rm -rf ./_out
