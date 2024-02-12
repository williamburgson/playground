package main

import (
	"context"
	"log"
	"time"
)

func watch(c *Cache) {
	for {
		select {
		case err := <-c.ErrChan:
			log.Printf("cache errored: %s", err)
			panic(err)
		case v := <-c.DelChan:
			log.Printf("%v removed from cache", v)
		}
	}
}

func main() {
	c := NewCache(5*time.Second, 1*time.Second)
	c.Add("foo", "bar")
	time.Sleep(1 * time.Second)
	c.Add("john", "doe")
	time.Sleep(1 * time.Second)
	c.Add("hello", "world")
	time.Sleep(1 * time.Second)
	ctx := context.Background()
	go func() {
		if err := c.Start(ctx); err != nil {
			log.Panicf("failed starting cache: %s", err)
		}
		if err := c.Wait(ctx); err != nil {
			log.Panicf("failed waiting cache to start: %s", err)
		}
	}()
	c.Add("newKey", "newValue")
	time.Sleep(1 * time.Second)
	c.Update("hello", "earth")
	watch(c)
}
