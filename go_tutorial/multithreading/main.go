package notmain

import (
	"fmt"
	"sync"
)

type Crawled struct {
	mu     sync.Mutex
	values map[string]string
}

func (c *Crawled) Update(key string, value string) {
	c.mu.Lock()
	c.values[key] = value
	c.mu.Unlock()
}

func Worker(wg *sync.WaitGroup, urls map[string]string, jobs, results chan string) {
	defer wg.Done()
	for j := range jobs {
		fmt.Println(j, "received")
		results <- urls[j]
	}
}

func main() {
	jobs, results := make(chan string), make(chan string)
	urls := map[string]string{
		"url1": "this is url1",
		"url2": "this is url2",
		"url3": "this is url3",
		"url4": "this is url4",
		"url5": "this is url5",
		"url6": "this is url6",
		"url7": "this is url7",
	}
	var wg sync.WaitGroup
	for k := range urls {
		go Worker(&wg, urls, jobs, results)
		jobs <- k
	}
	for i := 0; i < 7; i++ {
		fmt.Println(<-results)
	}
	wg.Wait()
}
