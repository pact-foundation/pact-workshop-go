include ./make/config.mk

install:
	@if [ ! -d pact/bin ]; then\
		echo "--- Installing Pact CLI dependencies";\
		curl -fsSL https://raw.githubusercontent.com/pact-foundation/pact-ruby-standalone/master/install.sh | bash;\
		go install github.com/pact-foundation/pact-go/v2@2.x.x;\
		pact-go -l DEBUG install;\
    fi
	go install github.com/pact-foundation/pact-go/v2@2.x.x;
	sudo $$HOME/go/bin/pact-go -l DEBUG install;

run-consumer:
	@go run consumer/client/cmd/main.go

run-provider:
	@go run provider/cmd/usersvc/main.go

deploy-consumer: install
	@echo "--- ✅ Checking if we can deploy consumer"
	@pact-broker can-i-deploy \
		--pacticipant $(CONSUMER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--to-environment production

deploy-provider: install
	@echo "--- ✅ Checking if we can deploy provider"
	@pact-broker can-i-deploy \
		--pacticipant $(PROVIDER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--to-environment production
record-deploy-consumer: install
	@echo "--- ✅ Recording deployment of consumer"
	pact-broker record-deployment \
		--pacticipant $(CONSUMER_NAME) \
		--broker-base-url ${PACT_BROKER_PROTO}://$(PACT_BROKER_URL) \
		--broker-username $(PACT_BROKER_USERNAME) \
		--broker-password $(PACT_BROKER_PASSWORD) \
		--version ${VERSION_COMMIT} \
		--environment production
record-deploy-provider: install
	@echo "--- ✅ Recording deployment of provider"
	pact-broker record-deployment \
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
consumer: install
	@echo "--- 🔨Running Consumer Pact tests "
	go test -tags=integration -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact'

provider: export PACT_TEST := true
provider: install
	@echo "--- 🔨Running Provider Pact tests "
	go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"

broker:
	docker-compose up -d
.PHONY: install deploy-consumer deploy-provider publish unit consumer provider
