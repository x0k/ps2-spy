package pubsub

type BufferedPublisher[E any] struct {
	Publisher[E]
	buffer []E
}

func NewBufferedPublisher[E any](pub Publisher[E], estimatedEventsCount int) *BufferedPublisher[E] {
	return &BufferedPublisher[E]{
		Publisher: pub,
		buffer:    make([]E, 0, estimatedEventsCount),
	}
}

func (p *BufferedPublisher[E]) Publish(event E) {
	p.buffer = append(p.buffer, event)
}

func (p *BufferedPublisher[E]) Flush() {
	for _, event := range p.buffer {
		p.Publisher.Publish(event)
	}
	p.buffer = p.buffer[:0]
}
