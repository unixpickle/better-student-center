package bsc

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func getNodeAttribute(node *html.Node, attribute string) string {
	lowerAttribute := strings.ToLower(attribute)
	for _, attr := range node.Attr {
		if strings.ToLower(attr.Key) == lowerAttribute {
			return attr.Val
		}
	}
	return ""
}

func nodeInnerText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	var res bytes.Buffer
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		res.WriteString(nodeInnerText(child))
	}
	return res.String()
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
		return nil, errors.New("wrong number of root elements")
	}

	root := parsed[0]

	var form loginFormInfo

	htmlForm, ok := scrape.Find(root, scrape.ByTag(atom.Form))
	if !ok {
		return nil, errors.New("no form element found")
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
		return nil, errors.New("no username field found")
	} else if form.passwordField == "" {
		return nil, errors.New("no password field found")
	}

	return &form, nil
}
