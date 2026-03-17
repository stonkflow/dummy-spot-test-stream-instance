package main

import (
	"fmt"
	arg "github.com/alexflint/go-arg"
)

var options struct {
	Workers int `arg:"--count,env:NUM_WORKERS"`
}

func main() {
	arg.MustParse(&options)
	fmt.Println("Workers:", options.Workers)
}
