package bsc

import (
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

type loginFormInfo struct {
	usernameField string
	passwordField string
	otherFields   url.Values
	action        string
}

// parseGenericLoginForm takes a login page and parses the first form it finds, treating it as the
// login form.
func parseGenericLoginForm(res *http.Response) (result *loginFormInfo, err error) {
	parsed, err := html.ParseFragment(res.Body, nil)
	if err != nil {
		return
	} else if len(parsed) != 1 {
		return nil, errors.New("Wrong number of root elements.")
	}

	root := parsed[0]

	var form loginFormInfo

	htmlForm, ok := scrape.Find(root, scrape.ByTag(atom.Form))
	if !ok {
		return nil, errors.New("No <form> found.")
	}

	if actionStr := getNodeAttribute(htmlForm, "action"); actionStr == "" {
		form.action = res.Request.URL.String()
	} else {
		actionURL, err := url.Parse(actionStr)
		if err != nil {
			return nil, err
		}
		if actionURL.Host == "" {
			actionURL.Host = res.Request.URL.Host
		}
		if actionURL.Scheme == "" {
			actionURL.Scheme = res.Request.URL.Scheme
		}
		if !path.IsAbs(actionURL.Path) {
			actionURL.Path = path.Join(res.Request.URL.Path, actionURL.Path)
		}
		form.action = actionURL.String()
	}

	inputs := scrape.FindAll(root, scrape.ByTag(atom.Input))
	form.otherFields = url.Values{}
	for _, input := range inputs {
		inputName := getNodeAttribute(input, "name")
		switch getNodeAttribute(input, "type") {
		case "text":
			form.usernameField = inputName
		case "password":
			form.passwordField = inputName
		default:
			form.otherFields.Add(inputName, getNodeAttribute(input, "value"))
		}
	}

	if form.usernameField == "" {
		return nil, errors.New("No username field found.")
	} else if form.passwordField == "" {
		return nil, errors.New("No password field found.")
	}

	return &form, nil
}
