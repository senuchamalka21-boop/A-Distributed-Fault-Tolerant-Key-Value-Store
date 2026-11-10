package main

import (
	"crypto/sha1"
	"encoding/binary"
	"sort"
	"strconv"
)

// Ring is a consistent hash ring.
type Ring struct {
	points []uint32          // sorted positions on the circle (0 .. 2^32-1)
	nodes  map[uint32]string // position -> node address
}

func NewRing() *Ring {
	return &Ring{nodes: make(map[uint32]string)}
}

// hashPos turns any string into a position on the ring.
func hashPos(s string) uint32 {
	h := sha1.Sum([]byte(s))
	return binary.BigEndian.Uint32(h[:4]) // take first 4 bytes -> uint32
}

// AddNode places `replicas` virtual points for this address onto the ring.
// 50 virtual nodes gives a much more even key distribution than 1.
func (r *Ring) AddNode(addr string) {
	for i := 0; i < 50; i++ {
		pos := hashPos(addr + "#" + strconv.Itoa(i))
		// avoid overwriting an existing point (rare collision)
		if _, exists := r.nodes[pos]; exists {
			continue
		}
		r.points = append(r.points, pos)
		r.nodes[pos] = addr
	}
	sort.Slice(r.points, func(i, j int) bool { return r.points[i] < r.points[j] })
}

// RemoveNode removes all virtual points belonging to addr.
func (r *Ring) RemoveNode(addr string) {
	newPoints := r.points[:0]
	for _, p := range r.points {
		if r.nodes[p] == addr {
			delete(r.nodes, p)
		} else {
			newPoints = append(newPoints, p)
		}
	}
	r.points = newPoints
}

// GetNode returns the node that owns `key` (first node clockwise from hash(key)).
func (r *Ring) GetNode(key string) string {
	if len(r.points) == 0 {
		return ""
	}
	pos := hashPos(key)
	idx := sort.Search(len(r.points), func(i int) bool { return r.points[i] >= pos })
	if idx == len(r.points) {
		idx = 0 // wrap around the ring
	}
	return r.nodes[r.points[idx]]
}

// GetN returns up to n distinct nodes starting at the owner and moving clockwise.
// Used for replication (owner + next replicas).
func (r *Ring) GetN(key string, n int) []string {
	if len(r.points) == 0 {
		return nil
	}
	pos := hashPos(key)
	idx := sort.Search(len(r.points), func(i int) bool { return r.points[i] >= pos })
	if idx == len(r.points) {
		idx = 0
	}

	seen := make(map[string]bool)
	result := []string{}
	for i := 0; i < len(r.points) && len(result) < n; i++ {
		node := r.nodes[r.points[(idx+i)%len(r.points)]]
		if !seen[node] {
			seen[node] = true
			result = append(result, node)
		}
	}
	return result
}