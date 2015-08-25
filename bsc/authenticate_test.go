package bsc

import (
	"fmt"
	"os"
	"testing"
)

var testAuthUsername string
var testAuthPassword string
var testAuthEngine UniversityEngine

func TestAuthenticate(t *testing.T) {
	badClient := NewClient(testAuthUsername, testAuthPassword+"POOP", testAuthEngine)
	if badClient.Authenticate() == nil {
		t.Error("Bad credentials returned successful result.")
	}
	goodClient := NewClient(testAuthUsername, testAuthPassword, testAuthEngine)
	if err := goodClient.Authenticate(); err != nil {
		t.Error("Login failed:", err)
	}
}

func TestMain(m *testing.M) {
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
		fmt.Fprintln(os.Stderr, "Unknown university: "+engineName)
		os.Exit(1)
	} else {
		testAuthEngine = engine
	}
	os.Exit(m.Run())
}

func showTestEnvVarHelp() {
	fmt.Fprintln(os.Stderr, "You must set the BSC_TEST_USERNAME, BSC_TEST_PASSWORD, and "+
		"BSC_TEST_UNIVERSITY enivorment variables before running these tests.")
	os.Exit(1)
}
