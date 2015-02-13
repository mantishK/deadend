package main

import (
	"fmt"
	"os"

	"github.com/mantishK/deadend"
)

func main() {
	args := os.Args
	sourceURL := args[1]

	brokenLinkChan := make(chan deadend.BrokenLinkMap, 100)
	deadend := deadend.NewDeadend(sourceURL)
	go deadend.Check(sourceURL, brokenLinkChan)
	fmt.Println("Fetching the broken links for " + sourceURL)
	for {
		select {
		case brokenLinkMap := <-brokenLinkChan:
			fmt.Println("Source: " + brokenLinkMap.Source)
			fmt.Println("Broken Link: " + brokenLinkMap.BrokenURL)
			fmt.Println("Status: ", brokenLinkMap.StatusCode)
			fmt.Println()
		}
	}
}
