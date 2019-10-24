TEST?=./...

include ./make/config.mk

install:
	@echo "--- Installing Pact CLI dependencies"
	curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash

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
	go test -v github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientUnit'

consumer:
	go test -v github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact'
	@echo "--- 🔨Running Provider Pact tests "

provider:
	@echo "--- 🔨Running Consumer Pact tests "
	go test -count=1 -tags=integration -v github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"

.PHONY: install deploy-consumer deploy-provider publish unit consumer provider
