package queue

import (
	"fmt"
	"sync"
)

// this is a stack ...
type Queue interface {
	Init()
	Add(k Key, v any)
	Remove(k Key)
	Pop() *Item
	Peep() *Item
}

type queue struct {
	mux   sync.RWMutex
	dict  map[Key]*Item
	items []*Item
}

type Key string

type Item struct {
	key     Key
	index   int
	content any
}

func (i *Item) KeyValue() (Key, any) {
	return i.key, i.content
}

func NewQueue() Queue {
	q := &queue{}
	q.Init()
	return q
}

func (q *queue) Init() {
	q.dict = map[Key]*Item{}
	q.items = []*Item{}
	q.mux = sync.RWMutex{}
}

func (q *queue) Add(k Key, v any) {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.add(&Item{key: k, content: v})
}

func (q *queue) Remove(k Key) {
	q.mux.Lock()
	defer q.mux.Unlock()
	q.remove(k)
}

func (q *queue) Pop() *Item {
	q.mux.Lock()
	defer q.mux.Unlock()
	return q.pop()
}

func (q *queue) Peep() *Item {
	q.mux.RLock()
	defer q.mux.RUnlock()
	return q.peep()
}

func (q *queue) pop() *Item {
	if len(q.items) == 0 {
		return nil
	}
	last := q.items[len(q.items)-1]
	delete(q.dict, last.key)
	dpcopy := q.items[:len(q.items)-1]
	q.items = dpcopy
	return last
}

func (q *queue) peep() *Item {
	if len(q.items) == 0 {
		return nil
	}
	return q.items[len(q.items)-1]
}

func (q *queue) add(v *Item) error {
	if _, ok := q.dict[v.key]; ok {
		return fmt.Errorf("%s is already present in q", v.key)
	}
	q.items = append(q.items, v)
	v.index = len(q.items) - 1
	q.dict[v.key] = v
	return nil
}

func (q *queue) remove(k Key) (*Item, error) {
	v, ok := q.dict[k]
	if !ok {
		return nil, fmt.Errorf("%s is not found in queue", k)
	}
	dpcopy := make([]*Item, len(q.items)-1)
	for i := 0; i < v.index; i++ {
		dpcopy[i] = q.items[i]
	}
	for i := v.index + 1; i < len(q.items); i++ {
		dpcopy[i-1] = q.items[i]
	}
	delete(q.dict, k)
	q.items = dpcopy
	return v, nil
}
