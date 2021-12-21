package main

import (
	"github.com/leep-frog/command/sourcerer"
	"github.com/leep-frog/todo"
)

func main() {
	sourcerer.Source(todo.CLI())
}
