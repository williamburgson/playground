package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Cache struct {
	cache           map[string]Value
	ttl             time.Duration
	refreshInterval time.Duration
	mu              sync.RWMutex
	waitChan        chan int
	ErrChan         chan error
	DelChan         chan Value
}

type Value struct {
	value   interface{}
	addedAt time.Time
}

func NewCache(ttl, interval time.Duration) *Cache {
	return &Cache{
		cache:           make(map[string]Value),
		ttl:             ttl,
		refreshInterval: interval,
		waitChan:        make(chan int, 1),
		ErrChan:         make(chan error, 1),
		DelChan:         make(chan Value, 1),
	}
}

func (c *Cache) Start(ctx context.Context) error {
	c.mu.RLock()
	c.waitChan <- -1
	c.mu.RUnlock()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(c.refreshInterval):
			if err := c.refresh(ctx); err != nil {
				c.ErrChan <- err
			}
		}
	}
}

func (c *Cache) Wait(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.waitChan:
			return nil
		}
	}
}

func (c *Cache) refresh(ctx context.Context) error {
	log.Print("refreshing cache ...\n")
	for k, v := range c.cache {
		age := time.Now().UTC().Sub(v.addedAt.UTC())
		log.Printf("refreshing %s, age %v, configured ttl %s, pending delete %v", k, age, c.ttl, age >= c.ttl)
		if age >= c.ttl {
			c.Delete(k)
		}
	}
	return nil
}

func (c *Cache) Add(k string, v interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[k] = Value{
		addedAt: time.Now(),
		value:   v,
	}
	log.Printf("new key value pair added to cache: %s, %v", k, v)
}

func (c *Cache) Delete(k string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.cache[k]; ok {
		delete(c.cache, k)
		c.DelChan <- v
	}
}

func (c *Cache) Update(k string, v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	old, ok := c.cache[k]
	if !ok {
		return fmt.Errorf("key %s not found in cache", k)
	}
	c.cache[k] = Value{
		addedAt: time.Now(),
		value:   v,
	}
	log.Printf("value for key %s updated from %v to %v", k, old, v)
	return nil
}

func (c *Cache) Get(k string) (interface{}, error) {
	c.mu.RLock()
	c.mu.RUnlock()
	v, ok := c.cache[k]
	if !ok {
		return nil, fmt.Errorf("key %s not found in cache", k)
	}
	return v, nil
}
