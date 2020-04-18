install:
	go install -v ${LDFLAGS}

test:
	@go test -v -cover ./...

cover:
	@go test -coverprofile cover.out
