all:	install

install:
	go install -i ./protoc-gen-go

test:
	go test -v ./...

clean:
	go clean ./...

nuke:
	go clean -i ./...
