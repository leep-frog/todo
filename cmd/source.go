package main

import (
	"os"

	"github.com/leep-frog/command/sourcerer"
	"github.com/leep-frog/todo"
)

func main() {
	os.Exit(sourcerer.Source(todo.CLI()))
}
