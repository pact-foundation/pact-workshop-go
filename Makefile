include ./make/config.mk

run-consumer:
	@go run consumer/client/cmd/main.go


.PHONY: install unit consumer run-consumer
