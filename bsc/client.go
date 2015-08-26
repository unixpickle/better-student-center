package bsc

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

var redirectionRejectedError = errors.New("a redirect occurred")

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
	httpClient := http.Client{Jar: jar, CheckRedirect: rejectRedirect}
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
	resp, err := c.client.Get(requestURL)
	if err != nil && !isRedirectError(err) {
		return nil, err
	} else if err == nil {
		return resp, nil
	}

	resp.Body.Close()

	fmt.Println("reauthenticating..., url is", requestURL)

	if err := c.Authenticate(); err != nil {
		fmt.Println("reauthenticate failed", err)
		return nil, err
	}

	fmt.Println("reauthenticated")

	resp, err = c.client.Get(requestURL)
	fmt.Println("got new error", err)
	if err != nil {
		return nil, err
	} else {
		return resp, nil
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

// isRedirectError returns true if an error is a redirectionRejectedError wrapped in url.Error.
func isRedirectError(err error) bool {
	if urlError, ok := err.(*url.Error); !ok {
		return false
	} else {
		return urlError.Err == redirectionRejectedError
	}
}

// rejectRedirect always returns an error.
func rejectRedirect(_ *http.Request, _ []*http.Request) error {
	return redirectionRejectedError
}
