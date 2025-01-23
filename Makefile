include ./make/config.mk

PACT_GO_VERSION=2.2.0
PACT_DOWNLOAD_DIR=/tmp
ifeq ($(OS),Windows_NT)
	PACT_DOWNLOAD_DIR=$$TMP
endif

install_cli:
	@if [ ! -d pact/bin ]; then\
		echo "--- Installing Pact CLI dependencies";\
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash;\
    fi

install:
	go install github.com/pact-foundation/pact-go/v2@v$(PACT_GO_VERSION)
	pact-go -l DEBUG install --libDir $(PACT_DOWNLOAD_DIR);

run-consumer:
	@go run consumer/client/cmd/main.go

run-provider:
	@go run provider/cmd/usersvc/main.go

deploy-consumer:
	@echo "--- ✅ Checking if we can deploy consumer"
	pact/bin/pact-broker can-i-deploy \
		--pacticipant $(CONSUMER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--to-environment production

deploy-provider:
	@echo "--- ✅ Checking if we can deploy provider"
	pact/bin/pact-broker can-i-deploy \
		--pacticipant $(PROVIDER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--to-environment production
record-deploy-consumer:
	@echo "--- ✅ Recording deployment of consumer"
	pact/bin/pact-broker record-deployment \
		--pacticipant $(CONSUMER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--environment production
record-deploy-provider:
	@echo "--- ✅ Recording deployment of provider"
	pact/bin/pact-broker record-deployment \
		--pacticipant $(PROVIDER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--environment production

publish:
	@echo "--- 📝 Publishing Pacts"
	pact/bin/pact-broker publish ${PWD}/pacts --consumer-app-version ${VERSION_COMMIT} --branch ${VERSION_BRANCH} \
		-b $(PACT_BROKER_PROTO)://$(PACT_BROKER_URL) -u ${PACT_BROKER_USERNAME} -p ${PACT_BROKER_PASSWORD}
	@echo
	@echo "Pact contract publishing complete!"
	@echo
	@echo "Head over to $(PACT_BROKER_PROTO)://$(PACT_BROKER_URL) and login with $(PACT_BROKER_USERNAME)/$(PACT_BROKER_PASSWORD)"
	@echo "to see your published contracts.	"
unit:
	@echo "--- 🔨Running Unit tests "
	go test -tags=unit -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientUnit'

consumer: export PACT_TEST := true
consumer:
	@echo "--- 🔨Running Consumer Pact tests "
	go test -tags=integration -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact' -v

provider: export PACT_TEST := true
provider:
	@echo "--- 🔨Running Provider Pact tests "
	go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v

broker:
	docker compose up -d
.PHONY: deploy-consumer deploy-provider publish unit consumer provider
