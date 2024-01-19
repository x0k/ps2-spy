package storage

type AbstractPublisher interface {
	Publish(event Event)
}

type TxPublisher struct {
	publisher AbstractPublisher
	buffer    []Event
}

func NewTxPublisher(publisher AbstractPublisher, estimatedEventsCount int) *TxPublisher {
	return &TxPublisher{
		publisher: publisher,
		buffer:    make([]Event, 0, estimatedEventsCount),
	}
}

func (b *TxPublisher) Publish(event Event) {
	b.buffer = append(b.buffer, event)
}

func (b *TxPublisher) Commit() {
	for _, event := range b.buffer {
		b.publisher.Publish(event)
	}
	b.buffer = nil
}
