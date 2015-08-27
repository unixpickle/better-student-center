package bsc

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"
)

var redirectionRejectedError = errors.New("redirect occurred")
var scheduleListViewPath string = "/EMPLOYEE/HRMS/c/SA_LEARNER_SERVICES.SSR_SSENRL_LIST.GBL" +
	"?Page=SSR_SSENRL_LIST"

// A Client makes requests to a University's Student Center.
type Client struct {
	// authLock ensures that no concurrent requests are made during the re-authentication process.
	// It also ensures that the client does not authenticate more than once concurrently.
	authLock sync.RWMutex

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
	return &Client{sync.RWMutex{}, httpClient, username, password, uni}
}

// Authenticate authenticates with the university's server.
//
// You should call this after creating a Client. However, if you do not, it will automatically be
// called after the first request fails.
func (c *Client) Authenticate() error {
	c.authLock.Lock()
	defer c.authLock.Unlock()
	return c.uni.Authenticate(c)
}

// FetchCurrentSchedule downloads the user's current schedule.
func (c *Client) FetchCurrentSchedule() ([]Course, error) {
	if resp, err := c.RequestPage(scheduleListViewPath); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		return ParseCurrentSchedule(resp.Body)
	}
}

// RequestPage requests a page relative to the PeopleSoft root. This will automatically
// re-authenticate if the session has timed out.
// If the request fails for any reason (including a redirect), the returned response is nil.
func (c *Client) RequestPage(page string) (*http.Response, error) {
	requestURL := c.uni.RootURL() + page
	c.authLock.RLock()
	resp, err := c.client.Get(requestURL)
	c.authLock.RUnlock()
	if err != nil && !isRedirectError(err) {
		return nil, err
	} else if err == nil {
		return resp, nil
	}

	resp.Body.Close()

	if err := c.Authenticate(); err != nil {
		return nil, err
	}

	c.authLock.RLock()
	resp, err = c.client.Get(requestURL)
	c.authLock.RUnlock()
	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return nil, err
	} else {
		return resp, nil
	}
}

// postGenericLoginForm uses parseGenericLoginForm on the given page and POSTs the username and
// password. It may fail at several points. If all is successful, it returns the result of the POST.
//
// Since this should only be called during authentication, it assumes that c.authLock is already
// locked in write mode.
//
// If the post results in a redirect, this may return a non-nil response with a non-nil error.
func (c *Client) postGenericLoginForm(authPageURL string) (*http.Response, error) {
	res, err := c.client.Get(authPageURL)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}

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
