
pre-commit:
	go mod tidy
	#fieldalignment -fix ./...
	golangci-lint run --issues-exit-code 1 -v "./..."