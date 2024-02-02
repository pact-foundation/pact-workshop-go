package provider

import (
	"fmt"
	l "log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pact-foundation/pact-go/v2/log"
	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
	"github.com/pact-foundation/pact-workshop-go/model"
	"github.com/pact-foundation/pact-workshop-go/provider/repository"
)

// The Provider verification
func TestPactProvider(t *testing.T) {
	log.SetLogLevel("DEBUG")

	go startInstrumentedProvider()

	verifier := provider.NewVerifier()

	// Verify the Provider - Branch-based Published Pacts for any known consumers
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		Provider:           "GoUserService",
		ProviderBaseURL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		ProviderBranch:     os.Getenv("VERSION_BRANCH"),
		FailIfNoPactsFound: false,
		PactFiles:          []string{filepath.FromSlash(fmt.Sprintf("%s/GoAdminService-GoUserService.json", os.Getenv("PACT_DIR")))},
		ProviderVersion:    os.Getenv("VERSION_COMMIT"),
		StateHandlers:      stateHandlers,
		RequestFilter:      fixBearerToken,
		BeforeEach: func() error {
			userRepository = sallyExists
			return nil
		},
	})

	if err != nil {
		t.Log(err)
	}
}

// Simulates the need to set a time-bound authorization token,
// such as an OAuth bearer token
func fixBearerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only set the correct bearer token, if one was provided in the first place
		if r.Header.Get("Authorization") != "" {
			r.Header.Set("Authorization", getAuthToken())
		}
		next.ServeHTTP(w, r)
	})
}

var stateHandlers = models.StateHandlers{
	"User sally exists": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
		userRepository = sallyExists
		return models.ProviderStateResponse{}, nil
	},
	"User sally does not exist": func(setup bool, s models.ProviderState) (models.ProviderStateResponse, error) {
		userRepository = sallyDoesNotExist
		return models.ProviderStateResponse{}, nil
	},
}

// Starts the provider API with hooks for provider states.
// This essentially mirrors the main.go file, with extra routes added.
func startInstrumentedProvider() {
	mux := GetHTTPHandler()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		l.Fatal(err)
	}
	defer ln.Close()

	l.Printf("API starting: port %d (%s)", port, ln.Addr())
	l.Printf("API terminating: %v", http.Serve(ln, mux))

}

// Configuration / Test Data
var port, _ = utils.GetFreePort()

// Provider States data sets
var sallyExists = &repository.UserRepository{
	Users: map[string]*model.User{
		"sally": {
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
		"sally": {
			FirstName: "Jean-Marie",
			LastName:  "de La Beaujardière😀😍",
			Username:  "sally",
			Type:      "blocked",
			ID:        10,
		},
	},
}
