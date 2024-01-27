package containers

import (
	"time"
)

type node[K comparable] struct {
	next, prev *node[K]
	key        K
	updatedAt  time.Time
}

type ExpirableList[K comparable] struct {
	nodes map[K]*node[K]
	head  *node[K]
	tail  *node[K]
}

func NewExpirableList[K comparable]() *ExpirableList[K] {
	return &ExpirableList[K]{
		nodes: make(map[K]*node[K]),
	}
}

func (l *ExpirableList[K]) Push(key K) {
	now := time.Now()
	if node, ok := l.nodes[key]; ok {
		if node == l.tail {
			return
		}
		// Should have next since it is not a tail
		node.next.prev = node.prev
		if node.prev != nil {
			node.prev.next = node.next
		}
		node.next = nil
		node.prev = l.tail
		node.updatedAt = now
		l.tail.next = node
		l.tail = node
		return
	}
	node := &node[K]{key: key, updatedAt: now}
	if l.tail == nil {
		l.head = node
		l.tail = node
	} else {
		l.tail.next = node
		node.prev = l.tail
		l.tail = node
	}
	l.nodes[key] = node
}

func (l *ExpirableList[K]) Remove(key K) {
	if node, ok := l.nodes[key]; ok {
		if node == l.head {
			l.head = node.next
		} else {
			node.prev.next = node.next
		}
		if node == l.tail {
			l.tail = node.prev
		} else {
			node.next.prev = node.prev
		}
		node.next = nil
		node.prev = nil
		delete(l.nodes, key)
	}
}

func (l *ExpirableList[K]) RemoveExpired(expirationTime time.Time, onRemove func(key K)) {
	curr := l.head
	for curr != nil && curr.updatedAt.Before(expirationTime) {
		k := curr.key
		onRemove(k)
		delete(l.nodes, k)
		curr = curr.next
	}
	if curr == nil {
		l.head = nil
		l.tail = nil
	} else {
		curr.prev = nil
		l.head = curr
	}
}
