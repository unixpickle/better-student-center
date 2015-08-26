package bsc

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
)

// A Client makes requests to a University's Student Center.
type Client struct {
	client   http.Client
	username string
	password string
	uni      UniversityEngine
}

// NewClient creates a new Client which authenticates with the supplied username, password, and
// UniversityEngine.
func NewClient(username, password string, uni UniversityEngine) *Client {
	jar, _ := cookiejar.New(nil)
	httpClient := http.Client{Jar: jar}
	return &Client{httpClient, username, password, uni}
}

// Authenticate authenticates with the university's server.
//
// You should call this after creating a Client. However, if you do not, it will automatically be
// called after the first request fails.
func (c *Client) Authenticate() error {
	return c.uni.Authenticate(c)
}

// FetchCourses downloads the user's current course list.
func (c *Client) FetchCourses() ([]Course, error) {
	if resp, err := c.RequestPage(enrolledCoursesPath); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		return ParseCourses(resp.Body)
	}
}

// RequestPage requests a page relative to the PeopleSoft root. This will automatically
// re-authenticate if the session has timed out.
func (c *Client) RequestPage(page string) (*http.Response, error) {
	requestURL := c.uni.RootURL() + page
	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	} else if resp.Request.URL.String() == requestURL {
		// NOTE: resp.Request will be different from the original request if a redirect occurred.
		// TODO: figure out if there's a nicer way to check for a redirect, or to use a
		// RoundTripper.
		return resp, nil
	}

	resp.Body.Close()

	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	resp, err = http.Get(requestURL)
	if err != nil {
		return nil, err
	} else if resp.Request.URL.String() == requestURL {
		return resp, nil
	} else {
		resp.Body.Close()
		return nil, errors.New("request redirected even after re-authentication")
	}
}

// postGenericLoginForm uses parseGenericLoginForm on the given page and POSTs the username and
// password. It may fail at several points. If all is successful, it returns the result of the POST.
func (c *Client) postGenericLoginForm(authPageURL string) (*http.Response, error) {
	res, err := c.client.Get(authPageURL)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	formInfo, err := parseGenericLoginForm(res)
	if err != nil {
		return nil, err
	}

	fields := formInfo.otherFields
	fields.Add(formInfo.usernameField, c.username)
	fields.Add(formInfo.passwordField, c.password)

	return c.client.PostForm(formInfo.action, fields)
}
