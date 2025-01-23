lint:
  golangci-lint run

test:
	go test -v -cover ./...

vendor:
	go mod vendor

yaegi_test:
	yaegi test -v .

clean:
	rm -rf ./vendor