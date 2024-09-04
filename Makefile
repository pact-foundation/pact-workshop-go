include ./make/config.mk

PACT_GO_VERSION=2.0.8
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

unit:
	@echo "--- 🔨Running Unit tests "
	go test -tags=unit -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientUnit' -v

consumer: export PACT_TEST := true
consumer:
	@echo "--- 🔨Running Consumer Pact tests "
	go test -tags=integration -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact' -v

provider: export PACT_TEST := true
provider:
	@echo "--- 🔨Running Provider Pact tests "
	go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v

.PHONY: install unit consumer provider run-provider run-consumer
