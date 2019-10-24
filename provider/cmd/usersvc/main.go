package main

import (
	"log"
	"net"
	"net/http"

	"github.com/pact-foundation/pact-workshop-go/provider"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/", provider.IsAuthenticated(provider.GetUser))
	mux.HandleFunc("/users/", provider.IsAuthenticated(provider.GetUsers))

	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Printf("API starting: port %d (%s)", 8080, ln.Addr())
	log.Printf("API terminating: %v", http.Serve(ln, mux))
}
