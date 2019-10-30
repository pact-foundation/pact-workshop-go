package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pact-foundation/pact-workshop-go/model"
)

// Client is our consumer interface to the Order API
type Client struct {
	BaseURL    *url.URL
	httpClient *http.Client
	Token      string
}

// WithToken applies a token to the
func (c *Client) WithToken(token string) *Client {
	c.Token = token
	return c
}

// GetUser gets a single user from the API
func (c *Client) GetUser(id int) (*model.User, error) {
	req, err := c.newRequest("GET", fmt.Sprintf("/user/%d", id), nil)
	if err != nil {
		return nil, err
	}
	var user model.User
	res, err := c.do(req, &user)

	if res != nil {
		switch res.StatusCode {
		case http.StatusNotFound:
			return nil, ErrNotFound
		case http.StatusUnauthorized:
			return nil, ErrUnauthorized
		}
	}

	if err != nil {
		return nil, ErrUnavailable
	}

	return &user, err

}

// GetUsers gets all users from the API
func (c *Client) GetUsers() ([]model.User, error) {
	req, err := c.newRequest("GET", "/users", nil)
	if err != nil {
		return nil, err
	}
	var users []model.User
	_, err = c.do(req, &users)

	return users, err
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	rel := &url.URL{Path: path}
	u := c.BaseURL.ResolveReference(rel)
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Admin Service")

	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	if c.httpClient == nil {
		c.httpClient = http.DefaultClient
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

var (
	// ErrNotFound represents a resource not found (404)
	ErrNotFound = errors.New("not found")

	// ErrUnauthorized represents a Forbidden (403)
	ErrUnauthorized = errors.New("unauthorized")

	ErrUnavailable = errors.New("api unavailable")
)
