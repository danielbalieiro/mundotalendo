package main

// This file must be named to sort before main.go lexicographically,
// ensuring its init() runs first and sets testMode before main's init().

func init() {
	testMode = true
}
