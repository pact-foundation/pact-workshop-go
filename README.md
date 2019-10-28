# Example project for the Pact workshop

This project has 2 components, a consumer project and a service provider as an Express API.

_NOTE: Each step is tied to, and must be run within, a git branch, allowing you to progress through each stage incrementally. For example, to move to step 2 run the following: `git checkout step2`_



## Scenario

There are two components in scope for our workshop.

1. Admin Service (Consumer). Does Admin-y things, and often needs to communicate to the User service. But really, it's just a label - it doesn't do much!
1. User Service (Provider). Provides useful things about a user, such as listing all users and getting the details of individuals.

For the purposes of this workshop, we won't implement any functionality of the Admin Service, except the bits that requires User information.

Whilst contract testing can be applied retrospectively to systems, we will follow the [consumer driven contracts](https://martinfowler.com/articles/consumerDrivenContracts.html) approach in this workshop - where a new consumer and provider are created in parallel to evolve a service over time.

## Step 1 - Simple Consumer calling Provider

We need to first create an HTTP client to make the calls to our provider service:

![Simple Consumer](diagrams/workshop_step1.png)

*NOTE*: even if the API client had been been graciously provided for us by our Provider Team, it doesn't mean that we shouldn't write contract tests - because the version of the client we have may not always be in sync with the deployed API - and also because we will write tests on the output appropriate to our specific needs.

This User Service expects a `user` path parameter, and then returns some simple json back:

![Sequence Diagram](diagrams/workshop_step1_class-sequence-diagram.png)

Great! We've created a client (see the `consumer/client` package).

We can run the client with `make run-consumer` - it should fail with an error, because the Provider is not running.

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
		assert.Equal(t, req.URL.String(), fmt.Sprintf("/user/%d", userID))
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

## Step 3 - Pact to the rescue

Let us add Pact to the project and write a consumer pact test for the `GET /users/:id` endpoint. Note how similar it looks to our unit test:

*consumer/client/client_pact_test.go:*

```go
	t.Run("the user exists", func(t *testing.T) {
		id := 10

		pact.
			AddInteraction().
			Given("User sally exists").
			UponReceiving("A request to login with user 'sally'").
			WithRequest(request{
				Method:  "GET",
				Path:    term("/users/10", "/user/[0-9]+"),
				Headers: headersWithToken,
			}).
			WillRespondWith(dsl.Response{
				Status:  200,
				Body:    dsl.Match(model.User{}),
				Headers: commonHeaders,
			})

		err := pact.Verify(func() error {
			user, err := client.WithToken("2019-01-01").GetUser(id)

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
```


![Test using Pact](diagrams/step3_pact.png)


This test starts a mock server a random port that actsio as our provider service. To get this to work we update the URL in the `Client` that we create, after initialising Pact.

Running this test still passes, but it creates a pact file which we can use to validate our assumptions on the provider side, and have conversation around.

```console
$ make consumer
```

Generated pact file (*pacts/our_little_consumer-our_provider.json*):

```json
{
  "consumer": {
    "name": "Our Little Consumer"
  },
  "provider": {
    "name": "Our Provider"
  },
  "interactions": [
    {
      "description": "a request for JSON data",
      "providerState": "data count > 0",
      "request": {
        "method": "GET",
        "path": "/provider",
        "query": "validDate=2017-06-12T08%3A04%3A24.387Z"
      },
      "response": {
        "status": 200,
        "headers": {
          "Content-Type": "application/json; charset=utf-8"
        },
        "body": {
          "test": "NO",
          "date": "2013-08-16T15:31:20+10:00",
          "count": 1000
        }
      }
    }
  ],
  "metadata": {
    "pactSpecification": {
      "version": "2.0.0"
    }
  }
}
```

# Step 4 - Verify the provider

![Pact Verification](diagrams/step4_pact.png)

We now need to validate the pact generated by the consumer is valid, by executing it against the running service provider, which should fail:

```console
npm run test:pact:provider

> pact-workshop-js@1.0.0 test:pact:provider /Users/mfellows/development/public/pact-workshop-js
> mocha provider/test/providerPact.spec.js



Provider Service listening on http://localhost:8081
  Pact Verification
    1) should validate the expectations of Our Little Consumer


  0 passing (538ms)
  1 failing

  1) Pact Verification should validate the expectations of Our Little Consumer:
     Error: Reading pact at pacts/our_little_consumer-our_provider.json

Verifying a pact between Our Little Consumer and Our Provider
  A request for json data
    with GET /provider?validDate=2017-06-12T08%3A29%3A09.261Z
      returns a response which
        has status code 200
        has a matching body (FAILED - 1)
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"

Failures:

  1) Verifying a pact between Our Little Consumer and Our Provider A request for json data with GET /provider?validDate=2017-06-12T08%3A29%3A09.261Z returns a response which has a matching body
     Failure/Error: expect(response_body).to match_term expected_response_body, diff_options

       Actual: {"test":"NO","validDate":"2017-06-12T08:31:42.460Z","count":100}

       @@ -1,4 +1,3 @@
        {
       -  "date": "2013-08-16T15:31:20+10:00",
       -  "count": 1000
       +  "count": 100
        }

       Key: - means "expected, but was not found".
            + means "actual, should not be found".
            Values where the expected matches the actual are not shown.

1 interaction, 1 failure
```

The test has failed for 2 reasons. Firstly, the count field has a different value to what was expected by the consumer.

Secondly, and more importantly, the consumer was expecting a `date` field while the provider generates a `validDate` field. Also, the date formats are different.

_NOTE_: We have separated the API provider into two components: one that provides a testable API and the other to start the actual service for local testing. You should now start the provider as follows:

```sh
node provider/providerService.js
```

# Step 5

Intentionally blank to align with the [JVM workshop](https://github.com/DiUS/pact-workshop-jvm/) steps

## Step 6 - Back to the client we go

Let's correct the consumer test to handle any integer for `count` and use the correct field for the `date`. Then we need to add a type matcher for `count` and change the field for the date to be `validDate`.

We can also add a date regular expression to make sure the `validDate` field is a valid date. This is important because we are parsing it.

The updated consumer test is now:

```js
const { somethingLike: like, term } = pact.Matchers

describe('Pact with Our Provider', () => {
  describe('given data count > 0', () => {
    describe('when a call to the Provider is made', () => {
      before(() => {
        return provider.setup()
          .then(() => {
            provider.addInteraction({
              uponReceiving: 'a request for JSON data',
              withRequest: {
                method: 'GET',
                path: '/provider',
                query: {
                  validDate: submissionDate
                }
              },
              willRespondWith: {
                status: 200,
                headers: {
                  'Content-Type': 'application/json; charset=utf-8'
                },
                body: {
                  test: 'NO',
                  validDate: term({generate: date, matcher: '\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}\\+\\d{2}:\\d{2}'}),
                  count: like(100)
                }
              }
            })
          })
      })
...
})
```

Running this test will fail until we fix the client. Here is the correct client function, which parses the date and formats it correctly:

```js
const fetchProviderData = (submissionDate) => {
  return request
    .get(`${API_ENDPOINT}/provider`)
    .query({validDate: submissionDate})
    .then((res) => {
      // Validate date
      if (res.body.validDate.match(/\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}/)) {
        return {
          count: res.body.count,
          date: moment(res.body.validDate, moment.ISO_8601).format('YYYY-MM-DDTHH:mm:ssZ')
        }
      } else {
        throw new Error('Invalid date format in response')
      }
    })
}
```

Now the test passes. But we still have a problem with the date format, which we must fix in the provider. Running the client now fails because of that.

```console
$ node consumer/consumer.js
Error: Invalid date format in response
    at request.get.query.then (consumer/client.js:20:15)
    at process._tickCallback (internal/process/next_tick.js:103:7)
```

## Step 7 - Verify the providers again

We need to run 'npm run test:pact:consumer' to publish the consumer pact file again. Then, running the provider verification tests we get the expected failure about the date format.

```
Failures:

...

@@ -1,4 +1,4 @@
{
-  "validDate": /\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}/
+  "validDate": "2017-06-12T10:28:12.904Z"
}
```

Lets fix the provider and then re-run the verification tests. Here is the corrected `/provider` resource:

```js
server.get('/provider', (req, res) => {
  const date = req.query.validDate

  res.json(
    {
      'test': 'NO',
      'validDate': moment(new Date(), moment.ISO_8601).format('YYYY-MM-DDTHH:mm:ssZ'),
      'count': 100
    }
  )
})
```

![Verification Passes](diagrams/step7_pact.png)

Running the verification against the providers now pass. Yay!

```console
$ npm run test:pact:provider

> pact-workshop-js@1.0.0 test:pact:provider /Users/mfellows/development/public/pact-workshop-js
> mocha provider/test/providerPact.spec.js



Provider Service listening on http://localhost:8081
  Pact Verification
Pact Verification Complete!
Reading pact at pacts/our_little_consumer-our_provider.json

Verifying a pact between Our Little Consumer and Our Provider
  A request for json data
    with GET /provider?validDate=2017-06-12T10%3A39%3A01.793Z
      returns a response which
        has status code 200
        has a matching body
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"

1 interaction, 0 failures



    ✓ should validate the expectations of Our Little Consumer (498ms)


  1 passing (503ms)
```

Our consumer also now works:

```console
$ node consumer/consumer.js
{ count: 100, date: '2017-06-12T20:40:51+10:00' }
```

## Step 8 - Test for the missing query parameter

In this step we are going to add a test for the case where the query parameter is missing or invalid. We do this by adding additional tests and expectations to the consumer pact test. Our client code needs to be modified slightly to be able to pass invalid dates in, and if the date parameter is null, don't include it in the request.

Here are the two additional tests:

*consumer/test/consumerPact.spec.js:*

```js
describe('and an invalid date is provided', () => {
  before(() => {
    return provider.addInteraction({
      uponReceiving: 'a request with an invalid date parameter',
      withRequest: {
        method: 'GET',
        path: '/provider',
        query: {
          validDate: 'This is not a date'
        }
      },
      willRespondWith: {
        status: 400,
        headers: {
          'Content-Type': 'application/json; charset=utf-8'
        },
        body: {'error': '"\'This is not a date\' is not a date"'}
      }
    })
  })

  it('can handle an invalid date parameter', (done) => {
    expect(fetchProviderData('This is not a date')).to.eventually.be.rejectedWith(Error).notify(done)
  })

  it('should validate the interactions and create a contract', () => {
    return provider.verify()
  })
})

describe('and no date is provided', () => {
  before(() => {
    return provider.addInteraction({
      uponReceiving: 'a request with a missing date parameter',
      withRequest: {
        method: 'GET',
        path: '/provider',
      },
      willRespondWith: {
        status: 400,
        headers: {
          'Content-Type': 'application/json; charset=utf-8'
        },
        body: {'error': '"validDate is required"'}
      }
    })
  })

  it('can handle missing date parameter', (done) => {
    expect(fetchProviderData(null)).to.eventually.be.rejectedWith(Error).notify(done)
  })

  it('should validate the interactions and create a contract', () => {
    return provider.verify()
  })
})
```

After running our specs, the pact file will have 2 new interactions.

*pacts/our_little_consumer-our_provider.json*:

```json
[
  {
    "description": "a request with an invalid date parameter",
    "request": {
      "method": "GET",
      "path": "/provider",
      "query": "validDate=This+is+not+a+date"
    },
    "response": {
      "status": 400,
      "headers": {
        "Content-Type": "application/json; charset=utf-8"
      },
      "body": {
        "error": "'This is not a date' is not a date"
      }
    }
  },
  {
    "description": "a request with a missing date parameter",
    "request": {
      "method": "GET",
      "path": "/provider"
    },
    "response": {
      "status": 400,
      "headers": {
        "Content-Type": "application/json; charset=utf-8"
      },
      "body": {
        "error": "validDate is required"
      }
    }
  }
]
```

## Step 9 - Verify the provider with the missing/invalid date query parameter

Let us run this updated pact file with Our Providers. We still get a 200 response as the provider doesn't yet do anything useful with the date.

Here is the provider test output:

```console
$ npm run test:pact:provider

> pact-workshop-js@1.0.0 test:pact:provider /Users/mfellows/development/public/pact-workshop-js
> mocha provider/test/providerPact.spec.js



Provider Service listening on http://localhost:8081
  Pact Verification
    1) should validate the expectations of Our Little Consumer


  0 passing (590ms)
  1 failing

  1) Pact Verification should validate the expectations of Our Little Consumer:
     Error: Reading pact at pacts/our_little_consumer-our_provider.json

Verifying a pact between Our Little Consumer and Our Provider
  A request for json data
    with GET /provider?validDate=2017-06-12T12%3A47%3A18.793Z
      returns a response which
        has status code 200
        has a matching body
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"
  A request with an invalid date parameter
    with GET /provider?validDate=This+is+not+a+date
      returns a response which
        has status code 400 (FAILED - 1)
        has a matching body (FAILED - 2)
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"
  A request with a missing date parameter
    with GET /provider
      returns a response which
        has status code 400 (FAILED - 3)
        has a matching body (FAILED - 4)
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"

Failures:

  1) Verifying a pact between Our Little Consumer and Our Provider A request with an invalid date parameter with GET /provider?validDate=This+is+not+a+date returns a response which has status code 400
     Failure/Error: expect(response_status).to eql expected_response_status

       expected: 400
            got: 200

       (compared using eql?)

  2) Verifying a pact between Our Little Consumer and Our Provider A request with an invalid date parameter with GET /provider?validDate=This+is+not+a+date returns a response which has a matching body
     Failure/Error: expect(response_body).to match_term expected_response_body, diff_options

       Actual: {"test":"NO","validDate":"2017-06-12T22:47:22+10:00","count":100}

       @@ -1,4 +1,3 @@
        {
       -  "error": "'This is not a date' is not a date"
        }

       Key: - means "expected, but was not found".
            + means "actual, should not be found".
            Values where the expected matches the actual are not shown.

  3) Verifying a pact between Our Little Consumer and Our Provider A request with a missing date parameter with GET /provider returns a response which has status code 400
     Failure/Error: expect(response_status).to eql expected_response_status

       expected: 400
            got: 200

       (compared using eql?)

  4) Verifying a pact between Our Little Consumer and Our Provider A request with a missing date parameter with GET /provider returns a response which has a matching body
     Failure/Error: expect(response_body).to match_term expected_response_body, diff_options

       Actual: {"test":"NO","validDate":"2017-06-12T22:47:22+10:00","count":100}

       @@ -1,4 +1,3 @@
        {
       -  "error": "validDate is required"
        }

       Key: - means "expected, but was not found".
            + means "actual, should not be found".
            Values where the expected matches the actual are not shown.

3 interactions, 2 failures
```

Time to update the providers to handle these cases.


## Step 10 - Update the providers to handle the missing/invalid query parameters

Let's fix Our Provider so it generate the correct responses for the query parameters.


The API resource gets updated to check if the parameter has been passed, and handle a date parse Error
if it is invalid. Two new Errors are thrown for these cases.

```js
server.get('/provider', (req, res) => {
  const validDate = req.query.validDate
  const dateRegex = /\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}/

  if (!validDate) {
    res.status(400)
    res.json({error: 'validDate is required'});
  } else if (!moment(validDate, moment.ISO_8601).isValid()) {
    res.status(400)
    res.json({error: `'${validDate}' is not a date`})
  }  else {
    res.json({
      'test': 'NO',
      'validDate': moment(new Date(), moment.ISO_8601).format('YYYY-MM-DDTHH:mm:ssZ'),
      'count': 100
    })
  }
})
```

Now running the `npm run test:pact:provider` will pass.

```console
$ npm run test:pact:provider

> pact-workshop-js@1.0.0 test:pact:provider /Users/mfellows/development/public/pact-workshop-js
> mocha provider/test/providerPact.spec.js



Provider Service listening on http://localhost:8081
  Pact Verification
Pact Verification Complete!
Reading pact at pacts/our_little_consumer-our_provider.json

Verifying a pact between Our Little Consumer and Our Provider
  A request for json data
    with GET /provider?validDate=2017-06-12T12%3A35%3A19.522Z
      returns a response which
        has status code 200
        has a matching body
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"
  A request with an invalid date parameter
    with GET /provider?validDate=This+is+not+a+date
      returns a response which
        has status code 400
        has a matching body
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"
  A request with a missing date parameter
    with GET /provider
      returns a response which
        has status code 400
        has a matching body
        includes headers
          "Content-Type" with value "application/json; charset=utf-8"

3 interactions, 0 failures



    ✓ should validate the expectations of Our Little Consumer (560ms)


  1 passing (566ms)
```

## Step 11 - Provider states

We have one final thing to test for. If the provider ever returns a count of zero, we will get a division by
zero error in our client. This is an important bit of information to add to our contract. Let us start with a
consumer test for this.

```js
  describe('given data count == 0', () => {
    describe('when a call to the Provider is made', () => {
      describe('and a valid date is provided', () => {
        before(() => {
          return provider.addInteraction({
            state: 'date count == 0',
            uponReceiving: 'a request for JSON data',
            withRequest: {
              method: 'GET',
              path: '/provider',
              query: { validDate: submissionDate }
            },
            willRespondWith: {
              status: 404,
              headers: {
                'Content-Type': 'application/json; charset=utf-8'
              },
              body: {
                test: 'NO',
                validDate: term({generate: date, matcher: dateRegex}),
                count: like(100)
              }
            }
          })
        })

        it('can handle missing data', (done) => {
          expect(fetchProviderData(submissionDate)).to.eventually.be.rejectedWith(Error).notify(done)
        })

        it('should validate the interactions and create a contract', () => {
          return provider.verify()
        })
      })
    })
  })
```

It is important to take note of the `state` property of the Interaction. There are two states: `"data count > 0"` and `"data count == 0"`. This is how we tell the provider during verification that it should prepare itself into a particular state, so as to illicit the response we are expecting.

This adds a new interaction to the pact file:

```json

  {
      "description": "a request for JSON data",
      "request": {
          "method": "GET",
          "path": "/provider.json",
          "query": "validDate=2017-05-22T13%3A34%3A41.515"
      },
      "response": {
          "status": 404
      },
      "providerState": "data count == 0"
  }

```

YOur Provider side verification will fail as it is not yet aware of these new 'states'.

## Step 12 - provider states for the providers

To be able to verify our providers, we need to be able to change the data that the provider returns. To do this, we need to instrument the running API with an extra [diagnostic endpoint](https://github.com/pact-foundation/pact-js#api-with-provider-states) to modify the data available to the API at runtime.

For our case, we are just going to use an in memory object to act as our persistence layer, but in a real project you would probably use a database.

Here is our data store:

```js
const dataStore = {
  count: 1000
}
```

Next, we update our API to use the value from the data store, and throw a 404 if there is no data.

```js
server.get('/provider', (req, res) => {
  const validDate = req.query.validDate
  const dateRegex = /\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\+\d{2}:\d{2}/

  if (!validDate) {
    res.status(400)
    res.json({error: 'validDate is required'});
  } else if (!moment(validDate, moment.ISO_8601).isValid()) {
    res.status(400)
    res.json({error: `'${validDate}' is not a date`})
  }  else {
    if (dataStore.count > 0) {
      res.json({
        'test': 'NO',
        'validDate': moment(new Date(), moment.ISO_8601).format('YYYY-MM-DDTHH:mm:ssZ'),
        'count': dataStore.count
      })
    } else {
      res.status(404)
      res.send()
    }
  }
})
```

Now we can change the data store value in our test based on the provider state to vary its behaviour.

Next we need to add a new endpoint to be able to manipulate this data store:

1. `/setup` to allow the Pact verification process to notify the API to switch to the new state

We do this by instrumenting the API in the _test code only_:

*provider/test/providerPact.spec.js*

```js
const { server, dataStore } = require('../provider.js')

// Set the current state
server.post('/setup', (req, res) => {
  switch (req.body.state) {
    case 'data count == 0':
      dataStore.count = 0
      break
    default:
      dataStore.count = 1000
  }

  res.end()
})
```

Lastly, we need to update our Pact configuration so that in knows how to find these new endpoits:

```js
let opts = {
  provider: 'Our Provider',
  providerBaseUrl: 'http://localhost:8081',
  providerStatesUrl: 'http://localhost:8081/states',
  providerStatesSetupUrl: 'http://localhost:8081/setup',
  pactUrls: [path.resolve(process.cwd(), './pacts/our_little_consumer-our_provider.json')]
}
```

Running the pact verification now passes:

```console
$ npm run test:pact:provider

> pact-workshop-js@1.0.0 test:pact:provider /Users/mfellows/development/public/pact-workshop-js
> mocha provider/test/providerPact.spec.js



Animal Profile Service listening on http://localhost:8081
  Pact Verification
Pact Verification Complete!
Reading pact at /Users/mfellows/development/public/pact-workshop-js/pacts/our_little_consumer-our_provider.json

Verifying a pact between Our Little Consumer and Our Provider
  Given date count > 0
    a request for JSON data
      with GET /provider?validDate=2017-06-12T13%3A14%3A49.745Z
        returns a response which
          has status code 200
          has a matching body
          includes headers
            "Content-Type" with value "application/json; charset=utf-8"
  Given date count > 0
    a request with an invalid date parameter
      with GET /provider?validDate=This+is+not+a+date
        returns a response which
          has status code 400
          has a matching body
          includes headers
            "Content-Type" with value "application/json; charset=utf-8"
  Given date count > 0
    a request with a missing date parameter
      with GET /provider
        returns a response which
          has status code 400
          has a matching body
          includes headers
            "Content-Type" with value "application/json; charset=utf-8"
  Given date count == 0
    a request for JSON data
      with GET /provider?validDate=2017-06-12T13%3A14%3A49.745Z
        returns a response which
          has status code 404
          includes headers
            "Content-Type" with value "application/json; charset=utf-8"

4 interactions, 0 failures



    ✓ should validate the expectations of Our Little Consumer (562ms)


  1 passing (567ms)

```

# Step 13 - Using a Pact Broker

We've been publishing our pacts from the consumer project by essentially sharing the file system with the provider. But this is not very manageable when you have multiple teams contributing to the code base, and pushing to CI. We can use a [Pact Broker](https://pact.dius.com.au) to do this instead.

Using a broker simplies the management of pacts and adds a number of useful features, including some safety enhancements for continuous delivery which we'll see shortly.

### Consumer

First, in the consumer project we need to tell Pact about our broker. We've created a small utility to push the pact files to the broker:

*consumer/test/publish:*

```groovy
const opts = {
  pactUrls: [path.resolve(__dirname, '../../pacts/our_little_consumer-our_provider.json')],
  pactBroker: 'https://test.pact.dius.com.au',
  pactBrokerUsername: 'dXfltyFMgNOFZAxr8io9wJ37iUpY42M',
  pactBrokerPassword: 'O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1',
  tags: ['prod', 'test'],
  consumerVersion: '1.0.0'
}

pact.publishPacts(opts)
```

You can run this with `test:pact:publish`:

```console
$ npm run test:pact:publish

> pact-workshop-js@1.0.0 test:pact:publish /Users/mfellows/development/public/pact-workshop-js
> node consumer/test/publish.js

Pact contract publishing complete!

Head over to https://test.pact.dius.com.au/ and login with
=> Username: dXfltyFMgNOFZAxr8io9wJ37iUpY42M
=> Password: O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1
to see your published contracts.
```

Have a browse around the broker and see your newly published contract!

### Provider

All we need to do for the provider is update where it finds its pacts from local URLs, to one from a broker:

```js
let opts = {
  provider: 'Our Provider',
  providerBaseUrl: 'http://localhost:8081',
  providerStatesUrl: 'http://localhost:8081/states',
  providerStatesSetupUrl: 'http://localhost:8081/setup',
  pactBrokerUrl: 'https://test.pact.dius.com.au/',
  tags: ['prod'],
  pactBrokerUsername: 'dXfltyFMgNOFZAxr8io9wJ37iUpY42M',
  pactBrokerPassword: 'O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1',
  publishVerificationResult: true,
  providerVersion: '1.1.0'
}
```

One thing you'll note in the output, is a message like this:

```console
Publishing verification result {"success":true,"providerApplicationVersion":"1.0.0"} to https://dXfltyFMgNOFZAxr8io9wJ37iUpY42M:*****@test.pact.dius.com.au/pacts/provider/Our%20Provider/consumer/Our%20Little%20Consumer/pact-version/1e37fb5393ea824ad898a5f12fbaa66af7ff3d3b/verification-results
```

This is a relatively new feature, but is very powerful. Called Verifications, it allows providers to report back the status of a verification to the broker. You'll get a quick view of the status of each consumer and provider on a nice dashboard. But, it is much more important than this!

With just a simple query to an API, we can quickly determine if a consumer is safe to release or not - the Broker will detect if any contracts for the consumer have changed and if so, have they been validated by each provider.

If something has changed, or it hasn't yet been validated by all downstream providers, then you should prevent any deployment going ahead. This is obviously really powerful for continuous delivery, which we alread had for providers.

Here is a simple cURL that will tell you if it's safe to release Our Little Consumer:

```sh
curl -s -u dXfltyFMgNOFZAxr8io9wJ37iUpY42M:O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1 "https://test.pact.dius.com.au/verification-results/consumer/Our%20Little%20Consumer/version/1.0.${USER}/latest" | jq .success
```

Or better yet, you can use our [CLI Tools](https://github.com/pact-foundation/pact-ruby-standalone/releases) to do the job, which are bundled as part of Pact JS:

```sh
npm run can-i-deploy:consumer
npm run can-i-deploy:provider
```

That's it - you're now a Pact pro!
