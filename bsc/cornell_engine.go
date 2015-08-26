package bsc

import "errors"

var cornellAuthURL string = "https://css.adminapps.cornell.edu/psc/cuselfservice/" +
	"EMPLOYEE/HRMS/c/SA_LEARNER_SERVICES.SSS_STUDENT_CENTER.GBL?" +
	"&FolderPath=PORTAL_ROOT_OBJECT.CO_EMPLOYEE_SELF_SERVICE.HC_" +
	"SSS_STUDENT_CENTER&IsFolder=false"
var cornellRootURL string = "https://css.adminapps.cornell.edu/psc/cuselfservice/"

// CornellEngine implements UniversityEngine for Cornell University's Student Center.
type CornellEngine struct{}

// Authenticate uses the CUWebLogin page to get a session.
func (_ CornellEngine) Authenticate(client *Client) error {
	res, err := client.client.Get(cornellAuthURL)
	if err == nil || !isRedirectError(err) {
		return errors.New("login page did not redirect")
	}
	fullURL := res.Header.Get("Location")

	res, err = client.postGenericLoginForm(fullURL)
	if res != nil {
		res.Body.Close()
	}

	// No redirects means that the login failed.
	if !isRedirectError(err) {
		return errors.New("login incorrect")
	}

	// Follow the redirects so that we get all the cookies from all the domains.
	// NOTE: I am cookie monster. Yummmm.
	for isRedirectError(err) {
		location := res.Header.Get("Location")
		res, err = client.client.Get(location)
	}
	return nil
}

// RootURL returns the root URL of Cornell's Student Center.
func (_ CornellEngine) RootURL() string {
	return cornellRootURL
}
