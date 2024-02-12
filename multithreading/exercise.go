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

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(ch chan crawlerArgs, result chan string, cache *crawlerCache, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	item := <-ch
	url, depth := item.url, item.depth
	if depth <= 0 {
		return
	}
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	result <- body
	cache.mu.Lock()
	cache.Update(url)
	cache.mu.Unlock()
	for _, u := range urls {
		_, exists := cache.cache[u]
		if !exists {
			continue
		}
		ch <- crawlerArgs{u, depth - 1}
		go Crawl(ch, result, cache, fetcher)
	}
	return
}

type crawlerArgs struct {
	url   string
	depth int
}

type crawlerCache struct {
	mu    sync.Mutex
	cache map[string]int
}

func (c *crawlerCache) Update(key string) {
	c.cache[key] = 1
}

func main() {
	ch := make(chan crawlerArgs)
	result := make(chan string)
	cache := crawlerCache{}
	cache.cache = map[string]int{}
	for f := range fetcher {
		go Crawl(ch, result, &cache, fetcher)
		ch <- crawlerArgs{f, len(fetcher[f].urls)}
		select {
		case r := <-result:
			fmt.Println(r)
		}
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
