# Pact Go workshop

## Step 10 - Request Filters on the Provider

Because our pact file has static data in it, our bearer token is now out of date, so when Pact verification passes it to the Provider we get a `401`. There are multiple ways to resolve this - mocking or stubbing out the authentication component is a common one. In our use case, we are going to use a process referred to as _Request Filtering_, using a `RequestFilter`. 

_NOTE_: This is an advanced concept and should be used carefully, as it has the potential to invalidate a contract by bypassing its constraints. See https://github.com/pact-foundation/pact-go#request-filtering for more details on this.

The approach we are going to take to inject the header is as follows:

1. If we receive any Authorization header, we override the incoming request with a valid (in time) Authorization header, and continue with whatever call was being made
1. If we don't recieve a header, we do nothing

_NOTE_: We are not considering the `403` scenario in this example.

Here is the request filter:

```go
// Simulates the neeed to set a time-bound authorization token,
// such as an OAuth bearer token
func fixBearerToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only set the correct bearer token, if one was provided in the first place
		if r.Header.Get("Authorization") != "" {
			r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", getAuthToken()))
		}
		next.ServeHTTP(w, r)
	})
}
```

We can now run the Provider tests

```
$ make provider

--- 🔨Running Provider Pact tests
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"
ok  	github.com/pact-foundation/pact-workshop-go/provider	22.138s
```

*Move on to [step 11](github.com/pact-foundation/pact-workshop-go/tree/master/steps/11)*