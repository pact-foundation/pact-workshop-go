package provider

import (
	"encoding/json"
	"net/http"

	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/pact-foundation/pact-workshop-go/model"
	"github.com/pact-foundation/pact-workshop-go/provider/repository"
)

var userRepository = &repository.UserRepository{
	Users: map[string]*model.User{
		"sally": &model.User{
			FirstName: "Jean-Marie",
			LastName:  "de La Beaujardière😀😍",
			Username:  "sally",
			Type:      "admin",
			ID:        10,
		},
	},
}

// IsAuthenticated checks for a correct bearer token
func WithCorrelationID(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		uuid := uuid.New()
		w.Header().Set("X-Api-Correlation-Id", uuid.String())
		h.ServeHTTP(w, r)
	}
}

// GetUser fetches a user if authenticated and exists
func GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Get username from path
	a := strings.Split(r.URL.Path, "/")
	id, _ := strconv.Atoi(a[len(a)-1])

	user, err := userRepository.ByID(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
		resBody, _ := json.Marshal(user)
		w.Write(resBody)
	}
}

// GetUsers fetches all users
func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resBody, _ := json.Marshal(userRepository.GetUsers())
	w.Write(resBody)
}

func commonMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return WithCorrelationID(f)
}

func GetHTTPHandler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/user/", commonMiddleware(GetUser))
	mux.HandleFunc("/users/", commonMiddleware(GetUsers))

	return mux
}
