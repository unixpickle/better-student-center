package bsc

import (
	"strings"

	"golang.org/x/net/html"
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
