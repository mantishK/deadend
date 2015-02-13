package deadend

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type Deadend struct {
	sourceURL string
	baseURL   string
	domainExt string
	visited   map[string]bool
}

type BrokenLinkMap struct {
	Source     string
	BrokenURL  string
	StatusCode int
}

func NewDeadend(sourceURL string) Deadend {
	urlRegex := regexp.MustCompile(`(www|http://www|http://).(([a-zA-Z0-9-]*).([a-z.]+))`)
	matchedURLs := urlRegex.FindAllStringSubmatch(sourceURL, -1)
	// Fetch the base url required to prepend to the relative links.
	baseURL := matchedURLs[0][0]
	// Fetch the domain name with its extension which is required to test if the links are internal or external
	domainExt := matchedURLs[0][2]
	// TODO: find a way to get rid of this map by writing it to the disk,
	// currently this is memory inefficient for large number of links.
	visited := make(map[string]bool)
	deadend := Deadend{sourceURL, baseURL, domainExt, visited}
	return deadend
}

func (deadend *Deadend) Check(sourceURL string, brokenLinkChan chan BrokenLinkMap) {
	deadend.checkURL(sourceURL, sourceURL, brokenLinkChan)
}

func (deadend *Deadend) checkURL(sourceURL, linkURL string, brokenLinkChan chan BrokenLinkMap) {
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
			go deadend.checkURL(linkURL, eachLinkURL, brokenLinkChan)
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
		if strings.Contains(linkURL, deadend.domainExt) {
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
		if !(strings.HasPrefix(link, "http") || strings.HasPrefix(link, "www")) {
			if strings.HasPrefix(link, "/") {
				link = deadend.baseURL + link
			} else {
				link = deadend.baseURL + "/" + link
			}
		}
		if !deadend.isVisited(link) {
			links[key] = link
		}
	}
	return links
}
