package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

// A Client makes requests to the Cornell Student Center.
type Client struct {
	client   http.Client
	username string
	password string
}

// NewClient creates a new Client which authenticates with the supplied username
// and password.
func NewClient(username, password string) *Client {
	jar, _ := cookiejar.New(nil)
	httpClient := http.Client{Jar: jar}
	return &Client{httpClient, username, password}
}

// Authenticate gets some sort of token to access the Student Center. In this
// case, the token is a session cookie.
//
// You should call this after creating a Client. However, if you do not, it will
// automatically be called after the first request fails.
func (c *Client) Authenticate() error {
	res, err := c.client.Get("https://studentcenter.cornell.edu/")
	if err != nil {
		return err
	}
	// TODO: this.
	return errors.New("Not yet implemented.")
}
