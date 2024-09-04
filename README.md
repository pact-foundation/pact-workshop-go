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

The correct endpoint should be `/user/:id`.

Move on to [step 5](//github.com/pact-foundation/pact-workshop-go/tree/step5)*

## Step 5 - Back to the client we go

![Pact Verification](diagrams/workshop_step5_pact.png)

Let's update the consumer test and client to hit the correct path, and run the provider verification also:

```
--- 🔨Running Consumer Pact tests 
go test -tags=integration -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact' -v
=== RUN   TestClientPact_GetUser
=== RUN   TestClientPact_GetUser/the_user_exists
2024-09-04T17:06:51.261564Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Received request GET /user/10
2024-09-04T17:06:51.262470Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Request matched, sending response
2024-09-04T17:06:51.263240Z  INFO ThreadId(02) pact_mock_server::mock_server: Writing pact out to '/Users/yousaf.nabi/dev/pact-foundation/pact-workshop-go/pacts/GoAdminService-GoUserService.json'
2024/09/04 18:06:51 [ERROR] failed to log to stdout: can't set logger (applying the logger failed, perhaps because one is applied already).
--- PASS: TestClientPact_GetUser (0.02s)
    --- PASS: TestClientPact_GetUser/the_user_exists (0.02s)
PASS
ok      github.com/pact-foundation/pact-workshop-go/consumer/client     0.071s
```

```
--- 🔨Running Provider Pact tests 
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v
=== RUN   TestPactProvider
2024/09/04 18:05:44 API starting: port 52781 ([::]:52781)
2024-09-04T17:05:45.133153Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024-09-04T17:05:45.133206Z  WARN ThreadId(11) pact_verifier::callback_executors: State Change ignored as there is no state change URL provided for interaction 
2024-09-04T17:05:45.133232Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:05:45.133294Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:52784/
2024-09-04T17:05:45.133297Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: None, body: Missing )
2024-09-04T17:05:45.136291Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 200, headers: Some({"date": ["Wed, 04 Sep 2024 17:05:45 GMT"], "content-length": ["109"], "content-type": ["application/json"], "x-api-correlation-id": ["1c0fa3a9-ebb7-4c21-ab01-345b882d0dc4"]}), body: Present(109 bytes, application/json) )
2024-09-04T17:05:45.137119Z  WARN ThreadId(11) pact_matching::metrics: 

Please note:
We are tracking events anonymously to gather important usage statistics like Pact version and operating system. To disable tracking, set the 'PACT_DO_NOT_TRACK' environment variable to 'true'.



Verifying a pact between GoAdminService and GoUserService

  A request to login with user 'sally' (0s loading, 166ms verification)
     Given User sally exists
    returns a response which
      has status code 200 (OK)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)


=== RUN   TestPactProvider/Provider_pact_verification
--- PASS: TestPactProvider (0.48s)
    --- PASS: TestPactProvider/Provider_pact_verification (0.00s)
PASS
ok      github.com/pact-foundation/pact-workshop-go/provider    0.532s
```

Yay - green ✅!

Move on to [step 6](//github.com/pact-foundation/pact-workshop-go/tree/step6)*

## Step 6 - Missing Users

We're now going to add another scenario - what happens when we make a call for a user that doesn't exist? We assume we'll get a `404`, because that is the obvious thing to do.

Let's write a test for this scenario, and then generate an updated pact file.

*consumer/client/client_pact_test.go*:
```go
	t.Run("the user does not exist", func(t *testing.T) {
		id := 10

		err = mockProvider.
			AddInteraction().
			Given("User sally does not exist").
			UponReceiving("A request to login with user 'sally'").
			WithRequestPathMatcher("GET", Regex("/user/"+strconv.Itoa(id), "/user/[0-9]+")).
			WillRespondWith(404, func(b *consumer.V2ResponseBuilder) {
				b.Header("Content-Type", Term("application/json", `application\/json`)).
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
				_, err := client.GetUser(id)
				assert.Equal(t, ErrNotFound, err)
				return nil
			})
			assert.NoError(t, err)
	})
```

Notice that our new test looks almost identical to our previous test, and only differs on the expectations of the _response_ - the HTTP request expectations are exactly the same.

```
$ make consumer

--- 🔨Running Consumer Pact tests 
go test -tags=integration -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact' -v
=== RUN   TestClientPact_GetUser
=== RUN   TestClientPact_GetUser/the_user_exists
2024-09-04T17:16:13.099939Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Received request GET /user/10
2024-09-04T17:16:13.101062Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Request matched, sending response
2024-09-04T17:16:13.101942Z  INFO ThreadId(02) pact_mock_server::mock_server: Writing pact out to '/Users/yousaf.nabi/dev/pact-foundation/pact-workshop-go/pacts/GoAdminService-GoUserService.json'
2024/09/04 18:16:13 [ERROR] failed to log to stdout: can't set logger (applying the logger failed, perhaps because one is applied already).
=== RUN   TestClientPact_GetUser/the_user_does_not_exist
2024-09-04T17:16:13.104166Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Received request GET /user/10
2024-09-04T17:16:13.104236Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Request matched, sending response
2024-09-04T17:16:13.104504Z  INFO ThreadId(03) pact_mock_server::mock_server: Writing pact out to '/Users/yousaf.nabi/dev/pact-foundation/pact-workshop-go/pacts/GoAdminService-GoUserService.json'
2024-09-04T17:16:13.104923Z  WARN ThreadId(03) pact_models::pact: Note: Existing pact is an older specification version (V2), and will be upgraded
2024/09/04 18:16:13 [ERROR] failed to log to stdout: can't set logger (applying the logger failed, perhaps because one is applied already).
--- PASS: TestClientPact_GetUser (0.03s)
    --- PASS: TestClientPact_GetUser/the_user_exists (0.02s)
    --- PASS: TestClientPact_GetUser/the_user_does_not_exist (0.00s)
PASS
ok      github.com/pact-foundation/pact-workshop-go/consumer/client     0.495s
```

What does our provider have to say about this new test:

```
--- 🔨Running Provider Pact tests 
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v
=== RUN   TestPactProvider
2024/09/04 18:16:16 API starting: port 52955 ([::]:52955)
2024-09-04T17:16:17.050635Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally does not exist' for 'A request to login with user 'sally''
2024-09-04T17:16:17.050689Z  WARN ThreadId(11) pact_verifier::callback_executors: State Change ignored as there is no state change URL provided for interaction 
2024-09-04T17:16:17.050715Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:16:17.050805Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:52958/
2024-09-04T17:16:17.050812Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: None, body: Missing )
2024-09-04T17:16:17.053140Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 200, headers: Some({"content-length": ["109"], "x-api-correlation-id": ["ff1ec2f1-c33e-4eaa-a4cf-108680275e0f"], "content-type": ["application/json"], "date": ["Wed, 04 Sep 2024 17:16:17 GMT"]}), body: Present(109 bytes, application/json) )
2024-09-04T17:16:17.200990Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024-09-04T17:16:17.201005Z  WARN ThreadId(11) pact_verifier::callback_executors: State Change ignored as there is no state change URL provided for interaction 
2024-09-04T17:16:17.201014Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:16:17.201045Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:52958/
2024-09-04T17:16:17.201047Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: None, body: Missing )
2024-09-04T17:16:17.202673Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 200, headers: Some({"content-length": ["109"], "content-type": ["application/json"], "date": ["Wed, 04 Sep 2024 17:16:17 GMT"], "x-api-correlation-id": ["14d99164-279d-4bba-b10e-93ac44d2113a"]}), body: Present(109 bytes, application/json) )
2024-09-04T17:16:17.203388Z  WARN ThreadId(11) pact_matching::metrics: 

Please note:
We are tracking events anonymously to gather important usage statistics like Pact version and operating system. To disable tracking, set the 'PACT_DO_NOT_TRACK' environment variable to 'true'.



Verifying a pact between GoAdminService and GoUserService

  A request to login with user 'sally' (0s loading, 167ms verification)
     Given User sally does not exist
    returns a response which
      has status code 404 (FAILED)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)

  A request to login with user 'sally' (0s loading, 149ms verification)
     Given User sally exists
    returns a response which
      has status code 200 (OK)
      includes headers
        "Content-Type" with value "application/json" (OK)
        "X-Api-Correlation-Id" with value "100" (OK)
      has a matching body (OK)


Failures:

1) Verifying a pact between GoAdminService and GoUserService Given User sally does not exist - A request to login with user 'sally'
    1.1) has status code 404
           expected 404 but was 200

There were 1 pact failures

=== RUN   TestPactProvider/Provider_pact_verification
    verifier.go:183: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
=== NAME  TestPactProvider
    user_service_test.go:36: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
--- FAIL: TestPactProvider (0.65s)
    --- FAIL: TestPactProvider/Provider_pact_verification (0.00s)
FAIL
FAIL    github.com/pact-foundation/pact-workshop-go/provider    0.698s
FAIL
make: *** [provider] Error 1
```

We expected this failure, because the user we are requesting does in fact exist! What we want to test for, is what happens if there is a different _state_ on the Provider. This is what is referred to as "Provider states", and how Pact gets around test ordering and related issues.

We could resolve this by updating our consumer test to use a known non-existent User, but it's worth understanding how Provider states work more generally.

*Move on to [step 7](//github.com/pact-foundation/pact-workshop-go/tree/step7)*

## Step 7 - Update our API to handle missing users

Our code already deals with missing users and sends a `404` response, however our test data fixture always has Sally (user `10`) in our database.

In this step, we will add a state handler (`StateHandlers`) to our Pact tests, which will update the state of our data store depending on which states.

States are invoked prior to the actual test function is invoked. You can see the full [lifecycle here](https://github.com/pact-foundation/pact-go#lifecycle-of-a-provider-verification).

We're going to add handlers for our two states - when Sally does and does not exist.

```go
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
```

Let's see how we go now:

```
$ make provider

--- 🔨Running Provider Pact tests 
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v
=== RUN   TestPactProvider
2024/09/04 18:23:27 API starting: port 53091 ([::]:53091)
2024-09-04T17:23:28.194956Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally does not exist' for 'A request to login with user 'sally''
2024/09/04 18:23:28 [INFO] executing state handler middleware
2024-09-04T17:23:28.340008Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:23:28.340096Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53094/
2024-09-04T17:23:28.340101Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: None, body: Missing )
2024-09-04T17:23:28.340969Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 404, headers: Some({"content-length": ["0"], "content-type": ["application/json"], "date": ["Wed, 04 Sep 2024 17:23:28 GMT"], "x-api-correlation-id": ["c958d383-d67b-4c95-b6a0-d48c77adf315"]}), body: Empty )
2024-09-04T17:23:28.341210Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User sally does not exist' for 'A request to login with user 'sally''
2024/09/04 18:23:28 [INFO] executing state handler middleware
2024-09-04T17:23:28.650479Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024/09/04 18:23:28 [INFO] executing state handler middleware
2024-09-04T17:23:28.804271Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:23:28.804312Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53094/
2024-09-04T17:23:28.804318Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: None, body: Missing )
2024-09-04T17:23:28.805289Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 200, headers: Some({"content-type": ["application/json"], "content-length": ["109"], "date": ["Wed, 04 Sep 2024 17:23:28 GMT"], "x-api-correlation-id": ["5b41178d-db3e-495e-8bd5-2fd525e5eef6"]}), body: Present(109 bytes, application/json) )
2024-09-04T17:23:28.805902Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024/09/04 18:23:28 [INFO] executing state handler middleware
2024-09-04T17:23:28.964743Z  WARN ThreadId(11) pact_matching::metrics: 

Please note:
We are tracking events anonymously to gather important usage statistics like Pact version and operating system. To disable tracking, set the 'PACT_DO_NOT_TRACK' environment variable to 'true'.



Verifying a pact between GoAdminService and GoUserService

  A request to login with user 'sally' (0s loading, 482ms verification)
     Given User sally does not exist
    returns a response which
      has status code 404 (OK)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)

  A request to login with user 'sally' (0s loading, 472ms verification)
     Given User sally exists
    returns a response which
      has status code 200 (OK)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)


=== RUN   TestPactProvider/Provider_pact_verification
--- PASS: TestPactProvider (1.28s)
    --- PASS: TestPactProvider/Provider_pact_verification (0.00s)
PASS
ok      github.com/pact-foundation/pact-workshop-go/provider    1.883s
```

*Move on to [step 8](//github.com/pact-foundation/pact-workshop-go/tree/step8)*

## Step 8 - Authorization

It turns out that not everyone should be able to use the API. After a discussion with the team, it was decided that a time-bound bearer token would suffice.

In the case a valid bearer token is not provided, we expect a `401`. Let's update the consumer test cases to pass the bearer token, and capture this new `401` scenario.

```
$ make consumer

--- 🔨Running Consumer Pact tests 
go test -tags=integration -count=1 github.com/pact-foundation/pact-workshop-go/consumer/client -run 'TestClientPact' -v
=== RUN   TestClientPact_GetUser
=== RUN   TestClientPact_GetUser/the_user_exists
2024-09-04T17:26:31.517409Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Received request GET /user/10
2024-09-04T17:26:31.518833Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Request matched, sending response
2024-09-04T17:26:31.520040Z  INFO ThreadId(02) pact_mock_server::mock_server: Writing pact out to '/Users/yousaf.nabi/dev/pact-foundation/pact-workshop-go/pacts/GoAdminService-GoUserService.json'
2024/09/04 18:26:31 [ERROR] failed to log to stdout: can't set logger (applying the logger failed, perhaps because one is applied already).
=== RUN   TestClientPact_GetUser/the_user_does_not_exist
2024-09-04T17:26:31.522280Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Received request GET /user/10
2024-09-04T17:26:31.522377Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Request matched, sending response
2024-09-04T17:26:31.522601Z  INFO ThreadId(03) pact_mock_server::mock_server: Writing pact out to '/Users/yousaf.nabi/dev/pact-foundation/pact-workshop-go/pacts/GoAdminService-GoUserService.json'
2024-09-04T17:26:31.522966Z  WARN ThreadId(03) pact_models::pact: Note: Existing pact is an older specification version (V2), and will be upgraded
2024/09/04 18:26:31 [ERROR] failed to log to stdout: can't set logger (applying the logger failed, perhaps because one is applied already).
=== RUN   TestClientPact_GetUser/not_authenticated
2024-09-04T17:26:31.524071Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Received request GET /user/10
2024-09-04T17:26:31.524138Z  INFO tokio-runtime-worker pact_mock_server::hyper_server: Request matched, sending response
2024-09-04T17:26:31.524262Z  INFO ThreadId(02) pact_mock_server::mock_server: Writing pact out to '/Users/yousaf.nabi/dev/pact-foundation/pact-workshop-go/pacts/GoAdminService-GoUserService.json'
2024-09-04T17:26:31.524448Z  WARN ThreadId(02) pact_models::pact: Note: Existing pact is an older specification version (V2), and will be upgraded
2024/09/04 18:26:31 [ERROR] failed to log to stdout: can't set logger (applying the logger failed, perhaps because one is applied already).
--- PASS: TestClientPact_GetUser (0.03s)
    --- PASS: TestClientPact_GetUser/the_user_exists (0.02s)
    --- PASS: TestClientPact_GetUser/the_user_does_not_exist (0.00s)
    --- PASS: TestClientPact_GetUser/not_authenticated (0.00s)
PASS
ok      github.com/pact-foundation/pact-workshop-go/consumer/client     0.473s
```

We should now have two interactions in our pact file.

Our verification now fails, as our consumer is sending a Bearer token that is not yet understood by our provider.

```
$ make provider

--- 🔨Running Provider Pact tests 
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v
=== RUN   TestPactProvider
2024/09/04 18:26:33 API starting: port 53171 ([::]:53171)
2024-09-04T17:26:33.691144Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User is not authenticated' for 'A request to login with user 'sally''
2024/09/04 18:26:33 [INFO] executing state handler middleware
2024/09/04 18:26:33 [WARN] no state handler found for state: User is not authenticated
2024-09-04T17:26:33.862810Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:26:33.862914Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53174/
2024-09-04T17:26:33.862919Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: None, body: Missing )
2024-09-04T17:26:33.863928Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 200, headers: Some({"x-api-correlation-id": ["df1e249c-ef86-4951-89cc-b36a865406b9"], "content-length": ["109"], "content-type": ["application/json"], "date": ["Wed, 04 Sep 2024 17:26:33 GMT"]}), body: Present(109 bytes, application/json) )
2024-09-04T17:26:33.864168Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User is not authenticated' for 'A request to login with user 'sally''
2024/09/04 18:26:34 [INFO] executing state handler middleware
2024/09/04 18:26:34 [WARN] no state handler found for state: User is not authenticated
2024-09-04T17:26:34.166077Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally does not exist' for 'A request to login with user 'sally''
2024/09/04 18:26:34 [INFO] executing state handler middleware
2024-09-04T17:26:34.321688Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:26:34.321728Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53174/
2024-09-04T17:26:34.321731Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: Some({"Authorization": ["Bearer 2019-01-01"]}), body: Missing )
2024-09-04T17:26:34.322490Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 404, headers: Some({"content-type": ["application/json"], "date": ["Wed, 04 Sep 2024 17:26:34 GMT"], "content-length": ["0"], "x-api-correlation-id": ["4f721ba2-d17f-4111-8652-6176f03db0c9"]}), body: Empty )
2024-09-04T17:26:34.322607Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User sally does not exist' for 'A request to login with user 'sally''
2024/09/04 18:26:34 [INFO] executing state handler middleware
2024-09-04T17:26:34.643400Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024/09/04 18:26:34 [INFO] executing state handler middleware
2024-09-04T17:26:34.848872Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:26:34.848925Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53174/
2024-09-04T17:26:34.848928Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: Some({"Authorization": ["Bearer 2019-01-01"]}), body: Missing )
2024-09-04T17:26:34.849747Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 200, headers: Some({"content-type": ["application/json"], "content-length": ["109"], "date": ["Wed, 04 Sep 2024 17:26:34 GMT"], "x-api-correlation-id": ["56978492-00c2-4f75-9697-d885caa660a0"]}), body: Present(109 bytes, application/json) )
2024-09-04T17:26:34.850493Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024/09/04 18:26:35 [INFO] executing state handler middleware
2024-09-04T17:26:35.004331Z  WARN ThreadId(11) pact_matching::metrics: 

Please note:
We are tracking events anonymously to gather important usage statistics like Pact version and operating system. To disable tracking, set the 'PACT_DO_NOT_TRACK' environment variable to 'true'.



Verifying a pact between GoAdminService and GoUserService

  A request to login with user 'sally' (0s loading, 498ms verification)
     Given User is not authenticated
    returns a response which
      has status code 401 (FAILED)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)

  A request to login with user 'sally' (0s loading, 478ms verification)
     Given User sally does not exist
    returns a response which
      has status code 404 (OK)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)

  A request to login with user 'sally' (0s loading, 511ms verification)
     Given User sally exists
    returns a response which
      has status code 200 (OK)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)


Failures:

1) Verifying a pact between GoAdminService and GoUserService Given User is not authenticated - A request to login with user 'sally'
    1.1) has status code 401
           expected 401 but was 200

There were 1 pact failures

=== RUN   TestPactProvider/Provider_pact_verification
    verifier.go:183: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
=== NAME  TestPactProvider
    user_service_test.go:44: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
--- FAIL: TestPactProvider (1.80s)
    --- FAIL: TestPactProvider/Provider_pact_verification (0.00s)
FAIL
FAIL    github.com/pact-foundation/pact-workshop-go/provider    2.291s
FAIL
make: *** [provider] Error 1
```

*Move on to [step 9](//github.com/pact-foundation/pact-workshop-go/tree/step9)*

## Step 9 - Implement authorisation on the provider

Like most tokens, our bearer token is going to be dependent on the date/time it was generated. For the purposes of our API, it's rather crude:

```go
func getAuthToken() string {
	return fmt.Sprintf("Bearer %s", time.Now().Format("2006-01-02T15:04"))
}
```

This means that a client must present an HTTP `Authorization` header that looks as follows:

```
Authorization: Bearer 2006-01-02T15:04
```

We have created a small middleware to wrap our functions and return a `401`:

```go
func IsAuthenticated(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == getAuthToken() {
			h.ServeHTTP(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
		}
	}
}
```

Let's test this out:

```
$ make provider

--- 🔨Running Provider Pact tests
--- 🔨Running Provider Pact tests 
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider" -v
=== RUN   TestPactProvider
2024/09/04 18:33:24 API starting: port 53320 ([::]:53320)
2024-09-04T17:33:24.470513Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User is not authenticated' for 'A request to login with user 'sally''
2024/09/04 18:33:24 [INFO] executing state handler middleware
2024/09/04 18:33:24 [WARN] no state handler found for state: User is not authenticated
2024-09-04T17:33:24.628543Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:33:24.628943Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53323/
2024-09-04T17:33:24.628947Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: None, body: Missing )
2024-09-04T17:33:24.629996Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 401, headers: Some({"content-type": ["application/json"], "content-length": ["0"], "date": ["Wed, 04 Sep 2024 17:33:24 GMT"], "x-api-correlation-id": ["412a28a3-b08f-4fce-b97d-65a9f90fe180"]}), body: Empty )
2024-09-04T17:33:24.630616Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User is not authenticated' for 'A request to login with user 'sally''
2024/09/04 18:33:24 [INFO] executing state handler middleware
2024/09/04 18:33:24 [WARN] no state handler found for state: User is not authenticated
2024-09-04T17:33:24.932360Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally does not exist' for 'A request to login with user 'sally''
2024/09/04 18:33:25 [INFO] executing state handler middleware
2024-09-04T17:33:25.085895Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:33:25.085991Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53323/
2024-09-04T17:33:25.085999Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: Some({"Authorization": ["Bearer 2019-01-01"]}), body: Missing )
2024-09-04T17:33:25.086912Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 401, headers: Some({"content-length": ["0"], "content-type": ["application/json"], "date": ["Wed, 04 Sep 2024 17:33:25 GMT"], "x-api-correlation-id": ["0fab9b01-3710-49a7-9a57-d826cc79ee4f"]}), body: Empty )
2024-09-04T17:33:25.087072Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User sally does not exist' for 'A request to login with user 'sally''
2024/09/04 18:33:25 [INFO] executing state handler middleware
2024-09-04T17:33:25.387705Z  INFO ThreadId(11) pact_verifier: Running setup provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024/09/04 18:33:25 [INFO] executing state handler middleware
2024-09-04T17:33:25.535975Z  INFO ThreadId(11) pact_verifier: Running provider verification for 'A request to login with user 'sally''
2024-09-04T17:33:25.536019Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request to provider at http://localhost:53323/
2024-09-04T17:33:25.536021Z  INFO ThreadId(11) pact_verifier::provider_client: Sending request HTTP Request ( method: GET, path: /user/10, query: None, headers: Some({"Authorization": ["Bearer 2019-01-01"]}), body: Missing )
2024-09-04T17:33:25.536676Z  INFO ThreadId(11) pact_verifier::provider_client: Received response: HTTP Response ( status: 401, headers: Some({"date": ["Wed, 04 Sep 2024 17:33:25 GMT"], "content-length": ["0"], "content-type": ["application/json"], "x-api-correlation-id": ["e6e7fc30-7fed-4b75-9294-c189890c7443"]}), body: Empty )
2024-09-04T17:33:25.536793Z  INFO ThreadId(11) pact_verifier: Running teardown provider state change handler 'User sally exists' for 'A request to login with user 'sally''
2024/09/04 18:33:25 [INFO] executing state handler middleware
2024-09-04T17:33:25.690099Z  WARN ThreadId(11) pact_matching::metrics: 

Please note:
We are tracking events anonymously to gather important usage statistics like Pact version and operating system. To disable tracking, set the 'PACT_DO_NOT_TRACK' environment variable to 'true'.



Verifying a pact between GoAdminService and GoUserService

  A request to login with user 'sally' (0s loading, 492ms verification)
     Given User is not authenticated
    returns a response which
      has status code 401 (OK)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (OK)

  A request to login with user 'sally' (0s loading, 454ms verification)
     Given User sally does not exist
    returns a response which
      has status code 404 (FAILED)
      includes headers
        "Content-Type" with value "application/json" (OK)
        "X-Api-Correlation-Id" with value "100" (OK)
      has a matching body (OK)

  A request to login with user 'sally' (0s loading, 449ms verification)
     Given User sally exists
    returns a response which
      has status code 200 (FAILED)
      includes headers
        "X-Api-Correlation-Id" with value "100" (OK)
        "Content-Type" with value "application/json" (OK)
      has a matching body (FAILED)


Failures:

1) Verifying a pact between GoAdminService and GoUserService Given User sally does not exist - A request to login with user 'sally'
    1.1) has status code 404
           expected 404 but was 401
2) Verifying a pact between GoAdminService and GoUserService Given User sally exists - A request to login with user 'sally'
    2.1) has a matching body
           / -> Expected body Present(98 bytes) but was empty
    2.2) has status code 200
           expected 200 but was 401

There were 2 pact failures

=== RUN   TestPactProvider/Provider_pact_verification
    verifier.go:183: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
=== NAME  TestPactProvider
    user_service_test.go:44: the verifier failed to successfully verify the pacts, this indicates an issue with the provider API
--- FAIL: TestPactProvider (1.74s)
    --- FAIL: TestPactProvider/Provider_pact_verification (0.00s)
FAIL
FAIL    github.com/pact-foundation/pact-workshop-go/provider    2.247s
FAIL
make: *** [provider] Error 1
```

Oh, dear. _Both_ tests are now failing. Can you understand why?

*Move on to [step 10](//github.com/pact-foundation/pact-workshop-go/tree/step10)*
