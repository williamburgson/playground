package queue

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	q := queue{}
	q.Init()
	q.add(&Item{key: "foo", content: "bar"})
	q.add(&Item{key: "john", content: "doe"})
	q.add(&Item{key: "fizz", content: "bazz"})
	assert.Len(t, q.items, 3)

	i := q.peep()
	assert.Equal(t, i.index, 2)
	assert.Equal(t, i.key, Key("fizz"))
	assert.Equal(t, i.content, "bazz")

	last := q.pop()
	assert.Equal(t, last.content, "bazz")
	assert.Len(t, q.items, 2)

	q.add(&Item{key: "key", content: "val"})
	q.add(&Item{key: "will", content: "wang"})
	q.add(&Item{key: "biz", content: "bam"})
	removed, err := q.remove("will")
	assert.NoError(t, err)
	assert.Equal(t, removed.content, "wang")
	keys := []Key{}
	for _, v := range q.items {
		keys = append(keys, v.key)
	}
	assert.Equal(t, keys, []Key{"foo", "john", "key", "biz"})
}
