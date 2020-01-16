# Pact Go workshop

## Step 7 - Update our API to handle missing users

Our code already deals with missing users and sends a `404` response, however our test data fixture always has Sally (user `10`) in our database.

In this step, we will add a state handler (`StateHandlers`) to our Pact tests, which will update the state of our data store depending on which states.

States are invoked prior to the actual test function is invoked. You can see the full [lifecycle here](https://github.com/pact-foundation/pact-go#lifecycle-of-a-provider-verification).

We're going to add handlers for our two states - when Sally does and does not exist.

```go
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
```

Let's see how we go now:

```
$ make provider

--- 🔨Running Provider Pact tests
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"
ok  	github.com/pact-foundation/pact-workshop-go/provider	22.138s
```

*Move on to [step 8](github.com/pact-foundation/pact-workshop-go/tree/master/steps/8)*