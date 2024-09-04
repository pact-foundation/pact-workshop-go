# Pact Go workshop

## Introduction
This workshop is aimed at demonstrating core features and benefits of contract testing with Pact.

Whilst contract testing can be applied retrospectively to systems, we will follow the [consumer driven contracts](https://martinfowler.com/articles/consumerDrivenContracts.html) approach in this workshop - where a new consumer and provider are created in parallel to evolve a service over time, especially where there is some uncertainty with what is to be built.

This workshop should take from 1 to 2 hours, depending on how deep you want to go into each topic.

**Workshop outline**:

- [step 1: **create consumer**](//github.com/pact-foundation/pact-workshop-go/tree/step1): Create our consumer before the Provider API even exists
- [step 2: **unit test**](//github.com/pact-foundation/pact-workshop-go/tree/step2): Write a unit test for our consumer
- [step 3: **pact test**](//github.com/pact-foundation/pact-workshop-go/tree/step3): Write a Pact test for our consumer
- [step 4: **pact verification**](//github.com/pact-foundation/pact-workshop-go/tree/step4): Verify the consumer pact with the Provider API
- [step 5: **fix consumer**](//github.com/pact-foundation/pact-workshop-go/tree/step5): Fix the consumer's bad assumptions about the Provider
- [step 6: **pact test**](//github.com/pact-foundation/pact-workshop-go/tree/step6): Write a pact test for `404` (missing User) in consumer
- [step 7: **provider states**](//github.com/pact-foundation/pact-workshop-go/tree/step7): Update API to handle `404` case
- [step 8: **pact test**](//github.com/pact-foundation/pact-workshop-go/tree/step8): Write a pact test for the `401` case
- [step 9: **pact test**](//github.com/pact-foundation/pact-workshop-go/tree/step9): Update API to handle `401` case
- [step 10: **request filters**](//github.com/pact-foundation/pact-workshop-go/tree/step10): Fix the provider to support the `401` case
- [step 11: **pact broker**](//github.com/pact-foundation/pact-workshop-go/tree/step11): Implement a broker workflow for integration with CI/CD

_NOTE: Each step is tied to, and must be run within, a git branch, allowing you to progress through each stage incrementally. For example, to move to step 2 run the following: `git checkout step2`_

## Learning objectives

If running this as a team workshop format, you may want to take a look through the [learning objectives](./LEARNING.md).

## Scenario

There are two components in scope for our workshop.

1. Admin Service (Consumer). Does Admin-y things, and often needs to communicate to the User service. But really, it's just a placeholder for a more useful consumer (e.g. a website or another microservice) - it doesn't do much!
1. User Service (Provider). Provides useful things about a user, such as listing all users and getting the details of individuals.

For the purposes of this workshop, we won't implement any functionality of the Admin Service, except the bits that require User information.

**Project Structure**

The key packages are shown below:

```sh
├── consumer	  # Contains the Admin Service Team (client) project
├── model         # Shared domain model
├── pact          # The directory of the Pact Standalone CLI
├── provider      # The User Service Team (provider) project
```

## Step 1 - Simple Consumer calling Provider

We need to first create an HTTP client to make the calls to our provider service:

![Simple Consumer](diagrams/workshop_step1.png)

*NOTE*: even if the API client had been been graciously provided for us by our Provider Team, it doesn't mean that we shouldn't write contract tests - because the version of the client we have may not always be in sync with the deployed API - and also because we will write tests on the output appropriate to our specific needs.

This User Service expects a `users` path parameter, and then returns some simple json back:

![Sequence Diagram](diagrams/workshop_step1_class-sequence-diagram.png)

You can see the client public interface we created in the `consumer/client` package:

```go

type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
}

// GetUser gets a single user from the API
func (c *Client) GetUser(id int) (*model.User, error) {
}
```

We can run the client with `make run-consumer` - it should fail with an error, because the Provider is not running.

*Move on to [step 2](//github.com/pact-foundation/pact-workshop-go/tree/step2): Write a unit test for our consumer*

## Step 2 - Client Tested but integration fails

Now lets create a basic test for our API client. We're going to check 2 things:

1. That our client code hit the expected endpoint
1. That the response is marshalled into a `User` object, with the correct ID

*consumer/client/client_test.go*

```go
func TestClientUnit_GetUser(t *testing.T) {
	userID := 10

	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, req.URL.String(), fmt.Sprintf("/users/%d", userID))
		user, _ := json.Marshal(model.User{
			FirstName: "Sally",
			LastName:  "McDougall",
			ID:        userID,
			Type:      "admin",
			Username:  "smcdougall",
		})
		rw.Write([]byte(user))
	}))
	defer server.Close()

	// Setup client
	u, _ := url.Parse(server.URL)
	client := &Client{
		BaseURL: u,
	}
	user, err := client.GetUser(userID)
	assert.NoError(t, err)

	// Assert basic fact
	assert.Equal(t, user.ID, userID)
}

```

![Unit Test With Mocked Response](diagrams/workshop_step2_unit_test.png)

Let's run this spec and see it all pass:

```
$ make unit

--- 🔨Running Unit tests
go test -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientUnit'
ok  	github.com/pact-foundation/pact-workshop-go/consumer/client	10.196s
```

Meanwhile, our provider team has started building out their API in parallel. Let's run our client against our provider (you'll need two terminals to do this):


```
# Terminal 1
$ make run-provider

2019/10/28 18:24:37 API starting: port 8080 ([::]:8080)

# Terminal 2
make run-consumer

2019/10/28 18:25:57 api unavailable
exit status 1
make: *** [run-consumer] Error 1

```

Doh! The Provider doesn't know about `/users/:id`. On closer inspection, the provider only knows about `/user/:id` and `/users`.

We need to have a conversation about what the endpoint should be, but first...

*Move on to [step 3](//github.com/pact-foundation/pact-workshop-go/tree/step3)*

## Step 3 - Pact to the rescue

Let's add Pact to the project. It comes in two parts.

- Installing pact-go cli
  - Required to install pact-go system libraries
- Adding pact-go as a dev dependency to your project.

Always check the installation instructions in the [docs](https://github.com/pact-foundation/pact-go/tree/master?tab=readme-ov-file#installation) for your platform.

The following command will install the pact-go CLI tool.

```console
$ go install github.com/pact-foundation/pact-go/v2
```

and we will use the pact-go CLI tool to install system libraries required by pact-go

```console
$ pact-go -l DEBUG install
```

You can use the provided make command to do this for you.

```console
$ make install
```

You can add `pact-go` to your project with the following

```console
$ go get github.com/pact-foundation/pact-go/v2
```

We can now write a consumer pact test for the `GET /users/:id` endpoint. 

Note how similar it looks to our unit test:

*consumer/client/client_pact_test.go:*

```go
	t.Run("the user exists", func(t *testing.T) {
		id := 10

		err = mockProvider.
			AddInteraction().
			Given("User sally exists").
			UponReceiving("A request to login with user 'sally'").
			WithRequestPathMatcher("GET", Regex("/users/"+strconv.Itoa(id), "/users/[0-9]+")).
			WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
				b.BodyMatch(model.User{}).
					Header("Content-Type", Term("application/json", `application\/json`)).
					Header("X-Api-Correlation-Id", Like("100"))
			}).
			ExecuteTest(t, func(config consumer.MockServerConfig) error {
				// Act: test our API client behaves correctly

				// Get the Pact mock server URL
				u, _ = url.Parse("http://" + config.Host + ":" + strconv.Itoa(config.Port))

				// Initialise the API client and point it at the Pact mock server
				client = &Client{
					BaseURL: u,
				}

				// // Execute the API client
				user, err := client.GetUser(id)

				// // Assert basic fact
				if user.ID != id {
					return fmt.Errorf("wanted user with ID %d but got %d", id, user.ID)
				}

				return err
			})

		assert.NoError(t, err)

	})
```


![Test using Pact](diagrams/workshop_step3_pact.png)


This test starts a Pact mock server on a random port that acts as our provider service. . We can access the update the `config.Host` & `config.Port` from `consumer.MockServerConfig` in the `ExecuteTest` block and pass these into the `Client` that we create, after initialising Pact. Pact will ensure our client makes the request stated in the interaction.

Running this test still passes, but it creates a pact file which we can use to validate our assumptions on the provider side, and have conversation around.

```console
$ make consumer
```

A pact file should have been generated in *pacts/GoAdminService-GoUserService.json*

*Move on to [step 4](//github.com/pact-foundation/pact-workshop-go/tree/step4)*

## Step 4 - Verify the provider

![Pact Verification](diagrams/workshop_step4_pact.png)

We now need to validate the pact generated by the consumer is valid, by executing it against the running service provider, which should fail:

```console
$ make provider

--- 🔨Running Provider Pact tests 
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v
=== RUN   TestPactProvider
2024/09/04 17:57:08 API starting: port 52668 ([::]:52668)
2024-09-04T16:57:08.543176Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024-09-04T16:57:08.543378Z  WARN ThreadId(11) pact_verifier::callback_executors: State Change ignored as there is no state change URL provided for interaction 
2024-09-04T16:57:08.543404Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T16:57:08.543486Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:52671/
2024-09-04T16:57:08.543488Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /users/10, query: None, headers: None, body: Missing )
2024-09-04T16:57:08.546904Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 200, headers: Some({"content-length": ["111"], "content-type": ["application/json"], "date": ["Wed, 04 Sep 2024 16:57:08 GMT"], "x-api-correlation-id": ["0d3aaf0f-027f-4170-b77b-e5a0b11b7f6c"]}), body: Present(111 bytes, application/json) )
2024-09-04T16:57:08.549372Z  WARN ThreadId(11) pact_matching::metrics: 

Please note:
We are tracking events anonymously to gather important usage statistics like Pact version and operating system. To disable tracking, set the 'PACT_DO_NOT_TRACK' environment variable to 'true'.



Verifying a pact between GoAdminService and GoUserService

  A request to login with user 'sally' (2ms loading, 167ms verification)
     Given User sally exists
    returns a response which
      has status code 200 (OK)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (FAILED)


Failures:

1) Verifying a pact between GoAdminService and GoUserService Given User sally exists - A request to login with user 'sally'
    1.1) has a matching body
           $ -> Type mismatch: Expected [{"firstName":"Jean-Marie","id":10,"lastName":"de La Beaujardière😀😍","type":"admin","username":"sally"}] (Array) to be the same type as {"firstName":"Sally","id":10,"lastName":"McSmiley Face😀😍","type":"admin","username":"sally"} (Object)

There were 1 pact failures

=== RUN   TestPactProvider/Provider_pact_verification
    verifier.go:183: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
=== NAME  TestPactProvider
    user_service_test.go:36: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
--- FAIL: TestPactProvider (0.49s)
    --- FAIL: TestPactProvider/Provider_pact_verification (0.00s)
FAIL
FAIL    github.com/pact-foundation/pact-workshop-go/provider    0.539s
FAIL
make: *** [provider] Error 1
```

The test has failed, as the expected path `/users/:id` is actually triggering the `/users` endpoint (which we don't need), and returning a _list_ of Users instead of a _single_ User. We incorrectly believed our provider was following a RESTful design, but the authors were too lazy to implement a better routing solution 🤷🏻‍♂️.

*Move on to [step 5](//github.com/pact-foundation/pact-workshop-go/tree/step5)*
