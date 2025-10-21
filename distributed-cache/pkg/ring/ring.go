package ring

import (
	"crypto/md5"
	"fmt"
	"sort"
	"sync"
)

type Ring struct {
	mu       sync.RWMutex
	nodes    map[uint32]string
	keys     []uint32
	replicas int
}

func New(replicas int) *Ring {
	return &Ring{
		nodes:    make(map[uint32]string),
		replicas: replicas,
	}
}

func (r *Ring) hash(key string) uint32 {
	h := md5.Sum([]byte(key))
	return uint32(h[0])<<24 | uint32(h[1])<<16 | uint32(h[2])<<8 | uint32(h[3])
}

func (r *Ring) AddNode(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := 0; i < r.replicas; i++ {
		virtualKey := fmt.Sprintf("%s:%d", addr, i)
		hash := r.hash(virtualKey)
		r.nodes[hash] = addr
	}

	r.rebuildKeys()
}

func (r *Ring) RemoveNode(addr string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	affectedKeys := []string{}
	for i := 0; i < r.replicas; i++ {
		virtualKey := fmt.Sprintf("%s:%d", addr, i)
		hash := r.hash(virtualKey)
		delete(r.nodes, hash)
	}

	r.rebuildKeys()
	return affectedKeys
}

func (r *Ring) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.keys) == 0 {
		return ""
	}

	hash := r.hash(key)
	idx := sort.Search(len(r.keys), func(i int) bool {
		return r.keys[i] >= hash
	})

	if idx == len(r.keys) {
		idx = 0
	}
	return r.nodes[r.keys[idx]]
}

func (r *Ring) GetNodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]bool)
	var nodes []string
	for _, node := range r.nodes {
		if !seen[node] {
			nodes = append(nodes, node)
			seen[node] = true
		}
	}
	return nodes
}

func (r *Ring) rebuildKeys() {
	r.keys = make([]uint32, 0, len(r.nodes))
	for k := range r.nodes {
		r.keys = append(r.keys, k)
	}
	sort.Slice(r.keys, func(i, j int) bool { return r.keys[i] < r.keys[j] })
}
