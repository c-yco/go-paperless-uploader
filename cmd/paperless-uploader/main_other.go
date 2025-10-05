//go:build !windows

package main

import "log"

func main() {
	if err := runApp(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}
