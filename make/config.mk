SHELL = "/bin/bash"

export PATH := $(PWD)/pact/bin:$(PATH)
export PATH
export PROVIDER_NAME = GoUserService
export CONSUMER_NAME = GoAdminService
export PACT_DIR = $(PWD)/pacts
export LOG_DIR = $(PWD)/log
export PACT_BROKER_PROTO = http
export PACT_BROKER_URL = localhost:8081
export PACT_BROKER_USERNAME = pact_workshop
export PACT_BROKER_PASSWORD = pact_workshop
export VERSION_COMMIT?=$(shell git rev-parse HEAD)
export VERSION_BRANCH?=$(shell git rev-parse --abbrev-ref HEAD)