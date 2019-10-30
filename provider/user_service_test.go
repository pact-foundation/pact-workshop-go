package provider

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
	"github.com/pact-foundation/pact-workshop-go/model"
	"github.com/pact-foundation/pact-workshop-go/provider/repository"
)

// The Provider verification
func TestPactProvider(t *testing.T) {
	go startInstrumentedProvider()

	pact := createPact()

	// Verify the Provider - Tag-based Published Pacts for any known consumers
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		Tags:               []string{"master"},
		FailIfNoPactsFound: false,
		PactURLs:           []string{filepath.FromSlash(fmt.Sprintf("%s/goadminservice-gouserservice.json", os.Getenv("PACT_DIR")))},
		ProviderVersion:    "1.0.0",
		StateHandlers:      stateHandlers,
	})

	if err != nil {
		t.Fatal(err)
	}

}

var stateHandlers = types.StateHandlers{
	"User sally exists": func() error {
		userRepository = sallyExists
		return nil
	},
	"User sally does not exist": func() error {
		userRepository = sallyDoesNotExist
		return nil
	},
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startInstrumentedProvider() {
	mux := GetHTTPHandler()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("API starting: port %d (%s)", port, ln.Addr())
	log.Printf("API terminating: %v", http.Serve(ln, mux))

}

// Configuration / Test Data
var dir, _ = os.Getwd()
var pactDir = fmt.Sprintf("%s/../../pacts", dir)
var logDir = fmt.Sprintf("%s/log", dir)
var port, _ = utils.GetFreePort()

// Provider States data sets
var sallyExists = &repository.UserRepository{
	Users: map[string]*model.User{
		"sally": &model.User{
			FirstName: "Jean-Marie",
			LastName:  "de La Beaujardière😀😍",
			Username:  "sally",
			Type:      "admin",
			ID:        10,
		},
	},
}

var sallyDoesNotExist = &repository.UserRepository{}

var sallyUnauthorized = &repository.UserRepository{
	Users: map[string]*model.User{
		"sally": &model.User{
			FirstName: "Jean-Marie",
			LastName:  "de La Beaujardière😀😍",
			Username:  "sally",
			Type:      "blocked",
			ID:        10,
		},
	},
}

// Setup the Pact client.
func createPact() dsl.Pact {
	return dsl.Pact{
		Provider:                 "GoUserService",
		LogDir:                   logDir,
		PactDir:                  pactDir,
		DisableToolValidityCheck: true,
		LogLevel:                 "INFO",
	}
}
