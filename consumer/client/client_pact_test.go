package client

import (
	"fmt"
	"os"
	"testing"

	"net/url"

	"github.com/pact-foundation/pact-go/dsl"
	"github.com/pact-foundation/pact-workshop-go/model"
)

var commonHeaders = dsl.MapMatcher{
	"Content-Type": term("application/json; charset=utf-8", `application\/json`),
}

func TestMain(m *testing.M) {
	// Setup Pact and related test stuff
	setup()

	// Run all the tests
	code := m.Run()

	// Shutdown the Mock Service and Write pact files to disk
	pact.WritePact()
	pact.Teardown()
	os.Exit(code)
}

func TestClientPact_GetUser(t *testing.T) {
	t.Run("when the user exists", func(t *testing.T) {
		id := 29

		pact.
			AddInteraction().
			Given("User sally exists").
			UponReceiving("A request to login with user 'sally'").
			WithRequest(request{
				Method: "GET",
				Path:   term("/user/10", "/user/[0-9]+"),
			}).
			WillRespondWith(dsl.Response{
				Status: 200,
				Body:   dsl.Match(model.User{}),
				Headers: dsl.MapMatcher{
					"X-Api-Correlation-Id": dsl.Like("100"),
					"Content-Type":         term("application/json; charset=utf-8", `application\/json`),
				},
			})

		err := pact.Verify(func() error {
			u, _ := url.Parse(fmt.Sprintf("http://localhost:%d", pact.Server.Port))
			client := &Client{
				BaseURL: u,
			}
			user, err := client.GetUser(id)

			fmt.Println(user)

			// Assert basic fact
			if user.ID != id {
				return fmt.Errorf("wanted user with ID %d but got %d", id, user.ID)
			}

			return err
		})

		if err != nil {
			t.Fatalf("Error on Verify: %v", err)
		}
	})
}
