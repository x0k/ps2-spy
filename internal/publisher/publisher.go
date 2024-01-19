package publisher

type Abstract[E any] interface {
	Publish(event E) error
}
