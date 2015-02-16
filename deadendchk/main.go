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
	doneChan := make(chan bool)
	deadend := deadend.NewDeadend(sourceURL)
	go deadend.Check(sourceURL, brokenLinkChan, doneChan)
	fmt.Println("Fetching the broken links for " + sourceURL)
	for {
		done := false
		select {
		case brokenLinkMap := <-brokenLinkChan:
			fmt.Println("Source: " + brokenLinkMap.Source)
			fmt.Println("Broken Link: " + brokenLinkMap.BrokenURL)
			fmt.Println("Status: ", brokenLinkMap.StatusCode)
			fmt.Println()
		case <-doneChan:
			done = true
			break
		}
		if done {
			break
		}
	}
}
