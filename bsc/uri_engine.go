package bsc

import "errors"

var uriAuthURL string = "https://appsaprod.uri.edu:9503/psp/sahrprod_m2/?cmd=login&languageCd=ENG"
var uriRootURL string = "https://appsaprod.uri.edu:9503/psc/sahrprod_m2"

// URIEngine implements UniversityEngine for the University of Rhode Island's Student Center.
type URIEngine struct{}

// Authenticate uses URI's e-campus login page to get a session.
func (_ URIEngine) Authenticate(client *Client) error {
	res, err := client.postGenericLoginForm(uriAuthURL)
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.Request.URL.Query().Get("errorCode") != "" {
		return errors.New("Login incorrect.")
	} else {
		return nil
	}
}

// RootURL returns the URL prefix that serves iframe content from URI's PeopleSoft system.
func (_ URIEngine) RootURL() string {
	return uriRootURL
}
