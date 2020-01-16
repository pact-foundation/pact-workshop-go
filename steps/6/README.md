# Pact Go workshop

## Step 6 - Missing Users

We're now going to add another scenario - what happens when we make a call for a user that doesn't exist? We assume we'll get a `404`, because that is the obvious thing to do. 

Let's write a test for this scenario, and then generate an updated pact file.

*consumer/client/client_pact_test.go*:
```go
	t.Run("the user does not exist", func(t *testing.T) {
		pact.
			AddInteraction().
			Given("User sally does not exist").
			UponReceiving("A request to login with user 'sally'").
			WithRequest(request{
				Method:  "GET",
				Path:    term("/user/10", "/user/[0-9]+"),
			}).
			WillRespondWith(dsl.Response{
				Status:  404,
				Headers: commonHeaders,
			})

		err := pact.Verify(func() error {
			_, err := client.GetUser(10)

			return err
		})

		assert.Equal(t, ErrNotFound, err)
  })
```

Notice that our new test looks almost identical to our previous test, and only differs on the expectations of the _response_ - the HTTP request expectations are exactly the same.

```
$ make consumer

go test github.com/pact-foundation/pact-workshop-go/consumer/client -run '^TestClientPact'
ok  	github.com/pact-foundation/pact-workshop-go/consumer/client	21.983s
```

What does our provider have to say about this new test:

```
--- 🔨Running Provider Pact tests
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"
2019/10/30 13:46:32 API starting: port 64046 ([::]:64046)
--- FAIL: TestPactProvider (11.56s)
    pact.go:416: Verifying a pact between GoAdminService and GoUserService Given User sally does not exist A request to login with user 'sally' with GET /user/10 returns a response which has status code 404

        expected: 404
             got: 200

        (compared using eql?)

    user_service_test.go:43: error verifying provider: exit status 1
```

We expected this failure, because the user we are requesing does in fact exist! What we want to test for, is what happens if there is a different _state_ on the Provider. This is what is referred to as "Provider states", and how Pact gets around test ordering and related issues.

We could resolve this by updating our consumer test to use a known non-existent User, but it's worth understanding how Provider states work more generally.

*Move on to [step 7](../7)*