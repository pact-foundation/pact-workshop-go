# Pact Go workshop

## Introduction
This workshop is aimed at demonstrating core features and benefits of contract testing with Pact.

Whilst contract testing can be applied retrospectively to systems, we will follow the [consumer driven contracts](https://martinfowler.com/articles/consumerDrivenContracts.html) approach in this workshop - where a new consumer and provider are created in parallel to evolve a service over time, especially where there is some uncertainty with what is to be built.

This workshop should take from 1 to 2 hours, depending on how deep you want to go into each topic.

**Workshop outline**:

- [step 1: **create consumer**](steps/1): Create our consumer before the Provider API even exists
- [step 2: **unit test**](steps/2): Write a unit test for our consumer
- [step 3: **pact test**](steps/3): Write a Pact test for our consumer
- [step 4: **pact verification**](steps/4): Verify the consumer pact with the Provider API
- [step 5: **fix consumer**](steps/5): Fix the consumer's bad assumptions about the Provider
- [step 6: **pact test**](steps/6): Write a pact test for `404` (missing User) in consumer
- [step 7: **provider states**](steps/7): Update API to handle `404` case
- [step 8: **pact test**](steps/8): Write a pact test for the `401` case
- [step 9: **pact test**](steps/9): Update API to handle `401` case
- [step 10: **request filters**](steps/10): Fix the provider to support the `401` case
- [step 11: **pact broker**](steps/11): Implement a broker workflow for integration with CI/CD

**All the steps are found in the [steps](steps/) sub-folder, feel free to skip around as needed!**

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
├── consumer		  # Contains the Admin Service Team (client) project
├── model         # Shared domain model
├── pact          # The directory of the Pact Standalone CLI
├── provider      # The User Service Team (provider) project
```

*Start with [step 1: **create consumer**](steps/1): Create our consumer before the Provider API even exists*
