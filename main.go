package main

import (
	"fmt"
	"os"

	"github.com/unixpickle/better-student-center/bsc"
)

func main() {
	engine, ok := bsc.EnginesByName[os.Getenv("BSC_TEST_UNIVERSITY")]
	if !ok {
		fmt.Fprintln(os.Stderr, "Unknown university: "+os.Getenv("BSC_TEST_UNIVERSITY"))
		os.Exit(1)
	}
	c := bsc.NewClient(os.Getenv("BSC_TEST_USERNAME"), os.Getenv("BSC_TEST_PASSWORD"), engine)
	if err := c.Authenticate(); err != nil {
		fmt.Fprintln(os.Stderr, "Authentication failed:", err)
		os.Exit(1)
	}
	fmt.Println("Authenticated")
	courses, err := c.FetchCurrentSchedule(false)
	fmt.Println(courses, err)
}
