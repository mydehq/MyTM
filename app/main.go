/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
/*
Package main is the entry point of the Go application.
In Go, the 'main' package is special: it's what tells the Go compiler "this is an executable program",
not just a library of code to be shared.
*/
package main

import (
	// This import path matches the 'module' name defined in 'go.mod'.
	// Even though it looks like a URL, Go looks for it locally first because it is part of the current module.
	// You don't need to push to GitHub for this to work!
	"github.com/mydehq/mytm/cmd"
)

/*
main function is the function that runs when you execute the program.
It's like the first line of your bash script.
*/
func main() {
	// We delegate all logic to the cmd package.
	// cmd.Execute() is a function we defined in 'cmd/root.go' to start the CLI.
	cmd.Execute()
}
