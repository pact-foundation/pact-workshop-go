package client

import (
	"fmt"
	"os"

	"github.com/pact-foundation/pact-go/dsl"
)

// Common test data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var pact dsl.Pact

// Aliases
var like = dsl.Like
var eachLike = dsl.EachLike
var term = dsl.Term

type s = dsl.String
type request = dsl.Request

func setup() {
	pact = createPact()

	// Proactively start service to get access to the port
	pact.Setup(true)
}

func createPact() dsl.Pact {
	return dsl.Pact{
		Consumer:                 os.Getenv("CONSUMER_NAME"),
		Provider:                 os.Getenv("PROVIDER_NAME"),
		LogDir:                   logDir,
		PactDir:                  pactDir,
		LogLevel:                 "INFO",
		DisableToolValidityCheck: true,
	}
}
