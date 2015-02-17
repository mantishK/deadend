package deadend

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

type Deadend struct {
	url     *url.URL
	visited map[string]bool
}

type BrokenLinkMap struct {
	Source     string
	BrokenURL  string
	StatusCode int
}

func NewDeadend(sourceURL string) (*Deadend, error) {
	url, err := url.Parse(sourceURL)
	if err != nil {
		return nil, err
	}

	// TODO: find a way to get rid of this map by writing it to the disk,
	// currently this is memory inefficient for large number of links.
	visited := make(map[string]bool)
	deadend := Deadend{url, visited}
	return &deadend, nil
}

func (deadend *Deadend) Check(sourceURL string, brokenLinkChan chan BrokenLinkMap, doneChan chan bool) {
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(1)
	deadend.checkURL(sourceURL, sourceURL, brokenLinkChan, waitGroup)
	waitGroup.Wait()
	doneChan <- true
}

func (deadend *Deadend) checkURL(sourceURL, linkURL string, brokenLinkChan chan BrokenLinkMap, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	if deadend.isVisited(linkURL) {
		return
	}
	status, body, err := deadend.testFor200(linkURL)
	if err != nil {
		return
	}
	if status != 200 {
		brokenLinkMap := BrokenLinkMap{sourceURL, linkURL, status}
		brokenLinkChan <- brokenLinkMap
	} else {
		deadend.markVisited(linkURL)
	}
	linkURLs := deadend.extractLinks(body)
	for _, eachLinkURL := range linkURLs {
		if eachLinkURL != "" {
			waitGroup.Add(1)
			go deadend.checkURL(linkURL, eachLinkURL, brokenLinkChan, waitGroup)
		}
	}
}

func (deadend *Deadend) markVisited(linkURL string) {
	deadend.visited[linkURL] = true
}

func (deadend *Deadend) isVisited(linkURL string) bool {
	if deadend.visited[linkURL] == true {
		return true
	}
	return false
}

func (deadend *Deadend) testFor200(linkURL string) (int, string, error) {
	resp, err := http.Get(linkURL)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body := ""
	if resp.StatusCode == 200 {

		//check if it is external url and only fetch the body if it is not
		if strings.Contains(linkURL, deadend.url.Host) {
			bodyArray, _ := ioutil.ReadAll(resp.Body)
			body = string(bodyArray)
		}
	}
	return resp.StatusCode, body, nil
}

func (deadend *Deadend) extractLinks(body string) []string {
	linkRegex := regexp.MustCompile(`a href=['"]?([^'" >]+)`)
	matchedArray := linkRegex.FindAllStringSubmatch(body, -1)
	links := make([]string, len(matchedArray))
	for key, matchedItem := range matchedArray {
		link := matchedItem[1]
		linkURL, err := url.Parse(link)
		if err != nil {
			continue
		}
		linkURL.Host = deadend.url.Host

		//check for mailto links, avoid them
		if strings.HasPrefix(linkURL.Path, "mailto:") {
			continue
		}

		if !deadend.isVisited(linkURL.String()) {
			links[key] = linkURL.String()
		}
	}
	return links
}
