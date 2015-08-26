package bsc

import (
	"errors"
	"net/url"
)

var uriAuthURL string = "https://appsaprod.uri.edu:9503/psp/sahrprod_m2/?cmd=login&languageCd=ENG"
var uriRootURL string = "https://appsaprod.uri.edu:9503/psc/sahrprod_m2/"

// URIEngine implements UniversityEngine for University of Rhode Island's eCampus
type URIEngine struct{}

// Authenticate uses URI's e-campus login page to get a session.
func (_ URIEngine) Authenticate(client *Client) error {
	// First fetch the login form to setup the session
	// This request will likely redirect to +="&" but that doesn't matter
	// However, if this breaks in the future it may be wise to start there
	res, err := client.client.Get(uriAuthURL)
	if res != nil {
		res.Body.Close()
	}

	res, err = client.postGenericLoginForm(uriAuthURL)
	if res == nil {
		return err
	}
	res.Body.Close()

	location := res.Header.Get("Location")
	if location == "" {
		return errors.New("login did not trigger any redirect")
	}

	parsed, err := url.Parse(location)
	if err != nil {
		return err
	}

	if parsed.Query().Get("errorCode") != "" {
		return errors.New("login incorrect")
	}
	return nil
}

// RootURL returns the URL prefix that serves iframe content from URI's PeopleSoft system
func (_ URIEngine) RootURL() string {
	return uriRootURL
}
