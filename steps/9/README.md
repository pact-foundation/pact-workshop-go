# Pact Go workshop

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
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"
2019/10/30 14:08:11 API starting: port 64214 ([::]:64214)
2019/10/30 14:08:22 [WARN] state handler not found for state: User is not authenticated
--- FAIL: TestPactProvider (11.55s)
    pact.go:416: Verifying a pact between GoAdminService and GoUserService Given User sally exists A request to login with user 'sally' with GET /user/10 returns a response which has status code 200

        expected: 200
             got: 401

        (compared using eql?)

    pact.go:416: Verifying a pact between GoAdminService and GoUserService Given User sally exists A request to login with user 'sally' with GET /user/10 returns a response which has a matching body
        757: unexpected token at 'null'
    pact.go:416: Verifying a pact between GoAdminService and GoUserService Given User sally does not exist A request to login with user 'sally' with GET /user/10 returns a response which has status code 404

        expected: 404
             got: 401
```

Oh, dear. _Both_ tests are now failing. Can you understand why?

*Move on to [step 10](github.com/pact-foundation/pact-workshop-go/tree/master/steps/10)*
