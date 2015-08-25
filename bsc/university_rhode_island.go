package bsc

type UniversityOfRhodeIsland struct{}

// URLPrefix returns the URL prefix that serves iframes.
func (_ UniversityOfRhodeIsland) URLPrefix() string {
	return "https://appsaprod.uri.edu:9503/psc/sahrprod_m2"
}
