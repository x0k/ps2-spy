package metrics

type stub struct{}

func NewStub() Metrics {
	return &stub{}
}
