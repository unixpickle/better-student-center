package bsc

import "errors"

var cornellAuthURL string = "https://studentcenter.cornell.edu/"

// Cornell implements UniversityEngine for Cornell University's Student Center.
type Cornell struct{}

// Authenticate uses the CUWebLogin page to authenticate the user by getting a session cookie.
func (_ Cornell) Authenticate(client *Client) error {
	res, err := client.postGenericLoginForm(cornellAuthURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.Request.URL.Path == "/loginAction" {
		return errors.New("Login incorrect.")
	} else {
		return nil
	}
}
