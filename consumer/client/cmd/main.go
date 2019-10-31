package main

import (
	"log"
	"net/url"

	"github.com/pact-foundation/pact-workshop-go/consumer/client"
)

func main() {
	u, _ := url.Parse("http://localhost:8080")
	client := &client.Client{
		BaseURL: u,
	}

	users, err := client.GetUser(10)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(users)
}
