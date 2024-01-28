package containers

import (
	"time"
)

type node[K comparable] struct {
	next, prev *node[K]
	key        K
	updatedAt  time.Time
}

type ExpirationQueue[K comparable] struct {
	nodes map[K]*node[K]
	head  *node[K]
	tail  *node[K]
}

func NewExpirationQueue[K comparable]() *ExpirationQueue[K] {
	return &ExpirationQueue[K]{
		nodes: make(map[K]*node[K]),
	}
}

func (l *ExpirationQueue[K]) Len() int {
	return len(l.nodes)
}

func (l *ExpirationQueue[K]) Has(key K) bool {
	_, ok := l.nodes[key]
	return ok
}

func (l *ExpirationQueue[K]) Push(key K) {
	now := time.Now()

	if node, ok := l.nodes[key]; ok {
		if node == l.tail {
			return
		}
		// Should have next since it is not a tail
		node.next.prev = node.prev
		if node == l.head {
			l.head = node.next
		} else {
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

func (l *ExpirationQueue[K]) Remove(key K) {
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

func (l *ExpirationQueue[K]) RemoveExpired(expirationTime time.Time, onRemove func(key K)) int {
	curr := l.head
	count := 0
	for curr != nil && curr.updatedAt.Before(expirationTime) {
		k := curr.key
		onRemove(k)
		delete(l.nodes, k)
		count++
		curr = curr.next
	}
	if curr == nil {
		l.head = nil
		l.tail = nil
	} else {
		if curr.prev != nil {
			curr.prev.next = nil
			curr.prev = nil
		}
		l.head = curr
	}
	return count
}
