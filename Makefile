include ./make/config.mk

run-consumer:
	@go run consumer/client/cmd/main.go

run-provider:
	@go run provider/cmd/usersvc/main.go

unit:
	@echo "--- 🔨Running Unit tests "
	go test -tags=unit -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientUnit' -v

.PHONY: install unit consumer provider run-provider run-consumer
