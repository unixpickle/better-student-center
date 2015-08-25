package bsc

// A UniversityEngine implements university-specific methods for their respective Student Centers.
type UniversityEngine interface {
	Authenticate(client *Client) error
	URLPrefix() string
}
