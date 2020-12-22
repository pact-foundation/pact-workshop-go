package main

import (
	"log"
	"net/url"
	"time"

	"github.com/pact-foundation/pact-workshop-go/consumer/client"
)

var token = time.Now().Format("2006-01-02T15:04")

func main() {
	u, _ := url.Parse("http://localhost:8080")
	client := &client.Client{
		BaseURL: u,
	}

	users, err := client.WithToken(token).GetUser(10)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(users)
}
