package population_tracker

import "time"

type Node[K comparable] struct {
	next, prev *Node[K]
	key        K
	updatedAt  time.Time
}

type List[K comparable] struct {
	nodes map[K]*Node[K]
	head  *Node[K]
	tail  *Node[K]
}

func (l *List[K]) Push(key K) {
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
	node := &Node[K]{key: key, updatedAt: now}
	if l.tail == nil {
		l.head = node
		l.tail = node
	} else {
		node.prev = l.tail
		l.tail = node
	}
	l.nodes[key] = node
}

func (l *List[K]) RemoveUntil(t time.Time, onRemove func(key K)) {
	curr := l.head
	for curr != nil && curr.updatedAt.Before(t) {
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
