package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
)

func main() {

	// Enable when running E2E/integration tests before a release
	version := os.Getenv("VERSION_COMMIT")

	// Publish the Pacts...
	p := dsl.Publisher{}

	fmt.Println("Publishing Pact files to broker", os.Getenv("PACT_DIR"), os.Getenv("PACT_BROKER_URL"))
	err := p.Publish(types.PublishRequest{
		PactURLs:        []string{filepath.FromSlash(fmt.Sprintf("%s/goadminservice-gouserservice.json", os.Getenv("PACT_DIR")))},
		PactBroker:      fmt.Sprintf("%s://%s", os.Getenv("PACT_BROKER_PROTO"), os.Getenv("PACT_BROKER_URL")),
		ConsumerVersion: version,
		Branch:          os.Getenv("VERSION_BRANCH"),
		BrokerUsername:  os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:  os.Getenv("PACT_BROKER_PASSWORD"),
	})

	if err != nil {
		log.Println("ERROR: ", err)
		os.Exit(1)
	}
}
