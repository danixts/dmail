.PHONY: fmt fmt-check vet lint test cover build check

COVER_THRESHOLD := 80

fmt:
	gofmt -w .

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Not formatted:"; echo "$$unformatted"; exit 1; \
	fi

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test -race -coverpkg=./... -coverprofile=coverage.out ./...

cover: test
	@total=$$(go tool cover -func=coverage.out | awk '/total:/ {print $$3}' | tr -d '%'); \
	echo "Total coverage: $$total%"; \
	awk "BEGIN{exit !($$total >= $(COVER_THRESHOLD))}" || \
		{ echo "Coverage $$total% is below $(COVER_THRESHOLD)%"; exit 1; }

build:
	go build ./...

check: fmt-check vet lint build cover
