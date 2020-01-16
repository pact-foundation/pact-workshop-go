# Pact Go workshop

## Step 11 - Using a Pact Broker

![Broker collaboration Workflow](diagrams/workshop_step10-broker.png)

We've been publishing our pacts from the consumer project by essentially sharing the file system with the provider. But this is not very manageable when you have multiple teams contributing to the code base, and pushing to CI. We can use a [Pact Broker](https://pactflow.io) to do this instead.

Using a broker simplies the management of pacts and adds a number of useful features, including some safety enhancements for continuous delivery which we'll see shortly.

In this workshop we will be using the open source Pact broker.

### Running the Pact Broker with docker-compose

In the root directory, run:

```console
docker-compose up
```


### Publish from consumer

First, in the consumer project we need to tell Pact about our broker. We've created a small utility to push the pact files to the broker:

```console
$ make publish

--- 📝 Publishing Pacts
go run consumer/client/pact/publish.go
Publishing Pact files to broker /Users/matthewfellows/development/pact-workshop-go/pacts test.pact.dius.com.au
2019/10/30 15:23:09 [INFO]
2019/10/30 15:23:09 [INFO] Tagging version 1.0.0 of GoAdminService as "master"
2019/10/30 15:23:09 [INFO] Publishing GoAdminService/GoUserService pact to pact broker at https://test.pact.dius.com.au
2019/10/30 15:23:09 [INFO] The given version of pact is already published. Overwriting...
2019/10/30 15:23:09 [INFO] The latest version of this pact can be accessed at the following URL (use this to configure the provider verification):
2019/10/30 15:23:09 [INFO] https://test.pact.dius.com.au/pacts/provider/GoUserService/consumer/GoAdminService/latest
2019/10/30 15:23:09 [INFO]
2019/10/30 15:23:09 [DEBUG] response from publish <nil>

Pact contract publishing complete!

Head over to https://test.pact.dius.com.au and login with
to see your published contracts.
```

Have a browse around the broker and see your newly published contract!

### Provider

All we need to do for the provider is update where it finds its pacts, from local URLs, to one from a broker.

```go
	_, err := pact.VerifyProvider(t, types.VerifyRequest{
		ProviderBaseURL:    fmt.Sprintf("http://127.0.0.1:%d", port),
		Tags:               []string{"master"},
		FailIfNoPactsFound: false,
		Verbose:            false,
		// Use this if you want to test without the Pact Broker
		// PactURLs:                   []string{filepath.FromSlash(fmt.Sprintf("%s/goadminservice-gouserservice.json", os.Getenv("PACT_DIR")))},
		BrokerURL:                  fmt.Sprintf("%s://%s", os.Getenv("PACT_BROKER_PROTO"), os.Getenv("PACT_BROKER_URL")),
		BrokerUsername:             os.Getenv("PACT_BROKER_USERNAME"),
		BrokerPassword:             os.Getenv("PACT_BROKER_PASSWORD"),
		PublishVerificationResults: true,
		ProviderVersion:            "1.0.0",
		StateHandlers:              stateHandlers,
		RequestFilter:              fixBearerToken,
  })
```

Let's run the provider verification one last time after this change:

```
$ make provider

--- 🔨Running Provider Pact tests
go test -count=1 -tags=integration github.com/pact-foundation/pact-workshop-go/provider -run "TestPactProvider"
ok  	github.com/pact-foundation/pact-workshop-go/provider	58.047s
```

As part of this process, the results of the verification - the outcome (boolean) and the detailed information about the failures at the interaction level - are published to the Broker also.

This is one of the Broker's more powerful features. Referred to as [Verifications](https://docs.pact.io/pact_broker/advanced_topics/provider_verification_results), it allows providers to report back the status of a verification to the broker. You'll get a quick view of the status of each consumer and provider on a nice dashboard. But, it is much more important than this!

With just a simple use of the `pact-broker` [can-i-deploy tool](https://docs.pact.io/pact_broker/advanced_topics/provider_verification_results) - the Broker will determine if a consumer or provider is safe to release to the specified environment.

You can run the `can-i-deploy` checks as follows:

```sh
$ make deploy-consumer

--- ✅ Checking if we can deploy consumer
Computer says yes \o/

CONSUMER       | C.VERSION | PROVIDER      | P.VERSION | SUCCESS?
---------------|-----------|---------------|-----------|---------
GoAdminService | 1.0.0     | GoUserService | 1.0.0     | true

All required verification results are published and successful


$ make deploy-provider

--- ✅ Checking if we can deploy provider
Computer says yes \o/

CONSUMER       | C.VERSION | PROVIDER      | P.VERSION | SUCCESS?
---------------|-----------|---------------|-----------|---------
GoAdminService | 1.0.0     | GoUserService | 1.0.0     | true

All required verification results are published and successful
```



That's it - you're now a Pact pro. Go build 🔨
