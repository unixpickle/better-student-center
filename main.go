package main

import (
	"fmt"

	"github.com/unixpickle/better-student-center/bsc"
)

func main() {
	c := bsc.NewClient("USERNAME", "PASSWORD", bsc.Cornell{})
	fmt.Println(c.Authenticate())
}
