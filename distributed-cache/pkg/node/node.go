package node

import (
	"sync"
	"time"
)

type Entry struct {
	Value     []byte
	ExpiresAt time.Time
}

type Node struct {
	data sync.Map
}

func New() *Node {
	return &Node{}
}

func (n *Node) Set(key string, value []byte, ttl time.Duration) {
	entry := &Entry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
	n.data.Store(key, entry)
}

func (n *Node) Get(key string) ([]byte, bool) {
	val, ok := n.data.Load(key)
	if !ok {
		return nil, false
	}

	entry := val.(*Entry)
	if time.Now().After(entry.ExpiresAt) {
		n.data.Delete(key)
		return nil, false
	}

	return entry.Value, true
}

func (n *Node) Delete(key string) {
	n.data.Delete(key)
}

func (n *Node) Cleanup() int {
	count := 0
	n.data.Range(func(key, value interface{}) bool {
		entry := value.(*Entry)
		if time.Now().After(entry.ExpiresAt) {
			n.data.Delete(key)
			count++
		}
		return true
	})
	return count
}
