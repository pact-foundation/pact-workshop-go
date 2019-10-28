TEST?=./...

include ./make/config.mk

install:
	@echo "--- Installing Pact CLI dependencies"
	curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash

run-consumer:
	@go run consumer/client/cmd/main.go

run-provider:
	@go run provider/cmd/usersvc/main.go

deploy-consumer:
	@echo "--- ✅ Checking if we can deploy consumer"
	@pact-broker can-i-deploy \
		--pacticipant $(CONSUMER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--latest

deploy-provider:
	@echo "--- ✅ Checking if we can deploy provider"
	@pact-broker can-i-deploy \
		--pacticipant $(PROVIDER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--latest

publish:
	@echo "--- 📝 Publishing Pacts"
	go run consumer/client/pact/publish.go

unit:
	@echo "--- 🔨Running Unit tests "
	go test -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientUnit'

consumer: export PACT_TEST := true
consumer:
	@echo "--- 🔨Running Consumer Pact tests "
	go test github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact'

provider: export PACT_TEST := true
provider:
	@echo "--- 🔨Running Provider Pact tests "
	go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"

.PHONY: install deploy-consumer deploy-provider publish unit consumer provider
