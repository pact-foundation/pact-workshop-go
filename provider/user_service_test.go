package provider

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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
		ProviderBaseURL:            fmt.Sprintf("http://127.0.0.1:%d", port),
		Tags:                       []string{"master"},
		FailIfNoPactsFound:         false,
		Verbose:                    false,
		BrokerURL:                  fmt.Sprintf("%s://%s", os.Getenv("PACT_BROKER_PROTO"), os.Getenv("PACT_BROKER_URL")),
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
		StateHandlers:              stateHandlers,
		RequestFilter:              fixBearerToken,
	})

	if err != nil {
		t.Fatal(err)
	}

}

var token = "" // token will be dynamic based on state etc.

// Simulates the neeed to set a time-bound authorization token,
// such as an OAuth bearer token
func fixBearerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("Authorization", token)
		next.ServeHTTP(w, r)
	})
}

var stateHandlers = types.StateHandlers{
	"User sally exists": func() error {
		userRepository = sallyExists
		token = fmt.Sprintf("Bearer %s", getAuthToken())
		return nil
	},
	"User sally is authenticated": func() error {
		userRepository = sallyExists
		token = fmt.Sprintf("Bearer %s", getAuthToken())
		return nil
	},
	"User sally is unauthorized": func() error {
		userRepository = sallyUnauthorized
		token = "invalid"

		return nil
	},
	"User sally is unauthenticated": func() error {
		userRepository = sallyUnauthorized
		token = "invalid"

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
	mux := http.NewServeMux()
	mux.HandleFunc("/user/", IsAuthenticated(GetUser))
	mux.HandleFunc("/users/", IsAuthenticated(GetUsers))

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
