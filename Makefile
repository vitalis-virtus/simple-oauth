format:
	gofmt -e -d .

run:
	go run cmd/main.go

.PHONY: format, run