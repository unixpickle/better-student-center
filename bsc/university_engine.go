package bsc

// A UniversityEngine implements university-specific methods for their respective Student Centers.
type UniversityEngine interface {
	Authenticate(client *Client) error
	RootURL() string
}

var EnginesByName map[string]UniversityEngine = map[string]UniversityEngine{
	"uri":     URIEngine{},
	"cornell": CornellEngine{},
}
