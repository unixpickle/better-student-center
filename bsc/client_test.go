package bsc

import (
	"fmt"
	"os"
	"testing"
)

var testOfflineOnly bool
var testAuthUsername string
var testAuthPassword string
var testAuthEngine UniversityEngine

func TestAuthenticate(t *testing.T) {
	if testOfflineOnly {
		t.Skip("offline tests do not cover authentication")
	}
	badClient := NewClient(testAuthUsername, testAuthPassword+"POOP", testAuthEngine)
	if badClient.Authenticate() == nil {
		t.Error("bad credentials returned successful result")
	}
	goodClient := NewClient(testAuthUsername, testAuthPassword, testAuthEngine)
	if err := goodClient.Authenticate(); err != nil {
		t.Error("login failed:", err)
	}
}

func TestFetchCurrentSchedule(t *testing.T) {
	if testOfflineOnly {
		t.Skip("offline tests do not cover course fetching")
	}
	c := NewClient(testAuthUsername, testAuthPassword, testAuthEngine)
	if err := c.Authenticate(); err != nil {
		t.Fatal("could not authenticate:", err)
	}
	if courses, err := c.FetchCurrentSchedule(false); err != nil {
		t.Error("failed to fetch courses:", err)
	} else if courses == nil || len(courses) == 0 {
		t.Error("course list is empty or nil")
	}
}

func TestMain(m *testing.M) {
	if os.Getenv("BSC_TEST_OFFLINE") != "" {
		testOfflineOnly = true
		os.Exit(m.Run())
	}

	testAuthUsername = os.Getenv("BSC_TEST_USERNAME")
	if testAuthUsername == "" {
		showTestEnvVarHelp()
	}
	testAuthPassword = os.Getenv("BSC_TEST_PASSWORD")
	if testAuthPassword == "" {
		showTestEnvVarHelp()
	}
	engineName := os.Getenv("BSC_TEST_UNIVERSITY")
	if engine, ok := EnginesByName[engineName]; !ok {
		fmt.Fprintln(os.Stderr, "unknown University: "+engineName)
		os.Exit(1)
	} else {
		testAuthEngine = engine
	}
	os.Exit(m.Run())
}

func showTestEnvVarHelp() {
	fmt.Fprintln(os.Stderr, "You must set the BSC_TEST_USERNAME, BSC_TEST_PASSWORD, and "+
		"BSC_TEST_UNIVERSITY enivorment variables before running these tests. Alternatively, "+
		"set BSC_TEST_OFFLINE to a non-empty string to disable online testing.")
	os.Exit(1)
}
