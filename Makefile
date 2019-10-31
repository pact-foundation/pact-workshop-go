TEST?=./...

include ./make/config.mk

install:
	@echo "--- Installing Pact CLI dependencies"
	curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash

run-consumer:
	@go run consumer/client/cmd/main.go

run-provider:
	@go run provider/cmd/usersvc/main.go

unit:
	@echo "--- 🔨Running Unit tests "
	go test -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientUnit'

.PHONY: install unit consumer  run-provider run-consumer
