package bsc

import "errors"

var cornellAuthURL string = "https://studentcenter.cornell.edu/"
var cornellRootURL string = "https://css.adminapps.cornell.edu/psc/cuselfservice/"

// CornellEngine implements UniversityEngine for Cornell University's Student Center.
type CornellEngine struct{}

// Authenticate uses the CUWebLogin page to get a session.
func (_ CornellEngine) Authenticate(client *Client) error {
	res, err := client.postGenericLoginForm(cornellAuthURL)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.Request.URL.Path == "/loginAction" {
		return errors.New("Login incorrect.")
	} else {
		return nil
	}
}

// RootURL returns the root URL of Cornell's Student Center.
func (_ CornellEngine) RootURL() string {
	return cornellRootURL
}
