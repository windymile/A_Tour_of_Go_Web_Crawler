package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	defer func() { 
		switch cc := <-crawlCounter; cc {
		case 1:
			close(crawlResults)
		default:
			crawlCounter <- (cc-1)
		}
	}()

	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		crawlResults <- fmt.Sprintf("%v", err)
		return
	}
	crawlResults <- fmt.Sprintf("found: %s %q", url, body)
	for _, u := range urls {
		var cu = <-crawlUrls
		if _, visited := cu[u]; visited {
			crawlUrls <- cu
		} else {
			crawlCounter <- (1 + <-crawlCounter)
			cu[u] = true
			crawlUrls <- cu
			go Crawl(u, depth-1, fetcher)
		}
	}
	return
}

var crawlResults = make(chan string, 10)
var crawlCounter = make(chan int, 1)
var crawlUrls = make(chan map[string]bool, 1)
func main() {
	crawlCounter <- 1
	crawlUrls <- map[string]bool{ "http://golang.org/": true }
	go Crawl("http://golang.org/", 4, fetcher)
	for msg := range crawlResults {
		fmt.Println(msg)
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

