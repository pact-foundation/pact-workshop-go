# Pact Go workshop

## Step 8 - Authorization

It turns out that not everyone should be able to use the API. After a discussion with the team, it was decided that a time-bound bearer token would suffice. 

In the case a valid bearer token is not provided, we expect a `401`. Let's update the consumer test cases to pass the bearer token, and capture this new `401` scenario.

```
$ make consumer

--- 🔨Running Consumer Pact tests
go test github.com/pact-foundation/pact-workshop-go/consumer/client -run '^TestClientPact'
ok  	github.com/pact-foundation/pact-workshop-go/consumer/client	21.983s
```

We should now have two interactions in our pact file.

Our verification now fails, as our consumer is sending a Bearer token that is not yet understood by our provider.

```
$ make provider

--- 🔨Running Provider Pact tests
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"
2019/10/30 13:28:47 API starting: port 63875 ([::]:63875)
2019/10/30 13:28:59 [WARN] state handler not found for state:
--- FAIL: TestPactProvider (11.54s)
    pact.go:416: Verifying a pact between GoAdminService and GoUserService A request to login with user 'sally' with GET /user/10 returns a response which has status code 401

        expected: 401
             got: 200

        (compared using eql?)

    user_service_test.go:43: error verifying provider: exit status 1
```

*Move on to [step 9](../9)*