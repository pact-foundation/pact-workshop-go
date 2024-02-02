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
	"github.com/pact-foundation/pact-go/v2/provider"
	"github.com/pact-foundation/pact-go/v2/utils"
)

// The Provider verification
func TestPactProvider(t *testing.T) {
	log.SetLogLevel("INFO")

	go startInstrumentedProvider()

	verifier := provider.NewVerifier()

	// Verify the Provider - From file
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		Provider:           "GoUserService",
		ProviderBaseURL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		ProviderBranch:     os.Getenv("VERSION_BRANCH"),
		FailIfNoPactsFound: false,
		PactFiles:          []string{filepath.FromSlash(fmt.Sprintf("%s/GoAdminService-GoUserService.json", os.Getenv("PACT_DIR")))},
		ProviderVersion:    os.Getenv("VERSION_COMMIT"),
	})

	if err != nil {
		t.Log(err)
	}
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
