package main

import (
	"fmt"
	"sync"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type Parallelizer struct {
	v  map[string]bool
	mu sync.Mutex
}

func (p *Parallelizer) Add(in string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.v[in] = true
}

func (p *Parallelizer) Get(in string) (bool, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	v, ok := p.v[in]
	return v, ok
}

var p Parallelizer = Parallelizer{v: make(map[string]bool)}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, quitChan chan bool) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	if depth <= 0 {
		quitChan <- true
		return
	}

	if _, ok := p.Get(url); !ok {
		p.Add(url)
	} else {
		quitChan <- true
		return
	}

	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		quitChan <- true
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	childrenQuit := make(chan bool)
	for _, u := range urls {
		go Crawl(u, depth-1, fetcher, childrenQuit)

		<-childrenQuit
	}
	quitChan <- true
}

func main() {
	childrenQuit := make(chan bool)
	go Crawl("https://golang.org/", 4, fetcher, childrenQuit)
	<-childrenQuit
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
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}
