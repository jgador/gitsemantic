package main

import (
	"context"
	"os"
)

func main() {
	os.Exit((&App{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}).Run(context.Background(), os.Args[1:]))
}
