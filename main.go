package main

import "fmt"

func main() {
	c := NewClient("USERNAME", "PASSWORD")
	fmt.Println(c.Authenticate())
}
