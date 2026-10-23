package module

type Starter interface {
	OnStart(hooks ...Runnable)
}

func (m *Module) OnStart(hooks ...Runnable) {
	m.onStart = append(m.onStart, hooks...)
}

type Stopper interface {
	OnStop(hooks ...Runnable)
}

func (m *Module) OnStop(hooks ...Runnable) {
	m.onStop = append(m.onStop, hooks...)
}
