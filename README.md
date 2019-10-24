# Example JS project for the Pact workshop

This project has 2 components, a consumer project and a service provider as an Express API.

_NOTE: Each step is tied to, and must be run within, a git branch, allowing you to progress through each stage incrementally. For example, to move to step 2 run the following: `git checkout step2`_

## Scenario

There are two components in scope for our workshop.

1. User Service (Consumer). Provides useful things about a user, including the current orders
1. Order Service (Provider). It's job is to be able to retreive and place Orders

For the purposes of this workshop, we won't implement any functionality of the User Service, except the bits that requires ordering information.


## Step 1 - Simple Consumer calling Provider

Given we have a client that needs to make a HTTP GET request to a provider service, and requires a response in JSON format.

![Simple Consumer](diagrams/workshop_step1.png)

The consumer client is quite simple and looks like this

_consumer/consumer.js:_

```js
request
  .get(`${API_ENDPOINT}/provider`)
  .query({ validDate: new Date().toISOString() })
  .then(res => {
    console.log(res.body)
  })
```

and the express provider resource

_provider/provider.js:_

```js
server.get('/provider/:', (req, res) => {
  const date = req.query.validDate

  res.json({
    test: 'NO',
    validDate: new Date().toISOString(),
    count: 100,
  })
})
```

This providers expects a `validDate` parameter in HTTP date format, and then return some simple json back.

![Sequence Diagram](diagrams/sequence_diagram.png)

Start the provider in a separate terminal:

```
$ node provider/provider.js
Provider Service listening on http://localhost:9123
```

Running the client works nicely.

```
$ node consumer/consumer.js
{ test: 'NO', validDate: '2017-06-12T06:25:42.392Z', count: 100 }
```


## Step 2 - Client Tested but integration fails

Now lets separate the API client (collaborator) that uses the data it gets back from the provider into its own module. Here is the updated client method that uses the returned data:

*consumer/client.js:*

```js
const fetchProviderData = () => {
  return request
    .get(`${API_ENDPOINT}/provider`)
    .query({validDate: new Date().toISOString()})
    .then((res) => {
      return {
        value: 100 / res.body.count,
        date: res.body.date
      }
    })
}
```

The consumer is now a lot simpler:

*consumer/consumer.js:*

```js
const client = require('./client')

client.fetchProviderData().then(response => console.log(response))
```

![Sequence 2](diagrams/step2_sequence_diagram.png)

Let's now test our updated client.

*consumer/test/consumer.spec.js:*

```js
describe('Consumer', () => {
  describe('when a call to the Provider is made', () => {
    const date = '2013-08-16T15:31:20+10:00'
    nock(API_HOST)
      .get('/provider')
      .query({validDate: /.*/})
      .reply(200, {
        test: 'NO',
        date: date,
        count: 1000
      })

    it('can process the JSON payload from the provider', done => {
      const {fetchProviderData} = require('../consumer')
      const response = fetchProviderData()

      expect(response).to.eventually.have.property('count', 1000)
      expect(response).to.eventually.have.property('date', date).notify(done)
    })
  })
})
```

![Unit Test With Mocked Response](diagrams/step2_unit_test.png)

Let's run this spec and see it all pass:

```
$ npm run test:consumer

> pact-workshop-js@1.0.0 test:consumer /Users/mfellows/development/public/pact-workshop-js
> mocha consumer/test/consumer.spec.js



  Consumer
    when a call to the Provider is made
      ✓ can process the JSON payload from the provider


  1 passing (24ms)

```

However, there is a problem with this integration point. Running the actual client against any of the providers results in problem!

```
$ node consumer/consumer.js
{ count: 100, date: undefined }
```

The provider returns a `validDate` while the consumer is
trying to use `date`, which will blow up when run for real even with the tests all passing. Here is where Pact comes in.

## Step 3 - Pact to the rescue

Let us add Pact to the project and write a consumer pact test.

*consumer/test/consumerPact.spec.js:*

```js
const provider = pact({
  consumer: 'Our Little Consumer',
  provider: 'Our Provider',
  port: API_PORT,
  log: path.resolve(process.cwd(), 'logs', 'pact.log'),
  dir: path.resolve(process.cwd(), 'pacts'),
  logLevel: LOG_LEVEL,
  spec: 2
})
const submissionDate = new Date().toISOString()
const date = '2013-08-16T15:31:20+10:00'
const expectedBody = {
  test: 'NO',
  date: date,
  count: 1000
}

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
                body: expectedBody
              }
            })
          })
      })

      it('can process the JSON payload from the provider', done => {
        const response = fetchProviderData(submissionDate)

        expect(response).to.eventually.have.property('count', 1000)
        expect(response).to.eventually.have.property('date', date).notify(done)
      })

      it('should validate the interactions and create a contract', () => {
        return provider.verify()
      })
    })

    // Write pact files to file
    after(() => {
      return provider.finalize()
    })
  })
})
```


![Test using Pact](diagrams/step3_pact.png)


This test starts a mock server on port 1234 that pretends to be our provider. To get this to work we needed to update
our consumer to pass in the URL of the provider. We also updated the `fetchProviderData` method to pass in the
query parameter.

Running this spec still passes, but it creates a pact file which we can use to validate our assumptions on the provider side.

```console
$ npm run "test:pact:consumer"

> pact-workshop-js@1.0.0 test:pact:consumer /Users/mfellows/development/public/pact-workshop-js
> mocha consumer/test/consumerPact.spec.js

  Pact with Our Provider
    when a call to the Provider is made
      when data count > 0
        ✓ can process the JSON payload from the provider
        ✓ should validate the interactions and create a contract


  2 passing (571ms)

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
