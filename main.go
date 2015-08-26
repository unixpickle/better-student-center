package main

import (
	"fmt"
	"os"

	"github.com/unixpickle/better-student-center/bsc"
)

func main() {
	engine, ok := bsc.EnginesByName[os.Getenv("BSC_TEST_UNIVERSITY")]
	if !ok {
		fmt.Fprintln(os.Stderr, "Unknown university: " + os.Getenv("BSC_TEST_UNIVERSITY"))
		os.Exit(1)
	}
	c := bsc.NewClient(os.Getenv("BSC_TEST_USERNAME"), os.Getenv("BSC_TEST_PASSWORD"), engine)
	courses, err := c.FetchCourses()
	fmt.Println(courses, err)
}
