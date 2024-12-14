package module

type PreStarter interface {
	PreStart(hooks ...Runnable)
}

func (m *Module) PreStart(hooks ...Runnable) {
	m.preStart = append(m.preStart, hooks...)
}

func (m *Module) PreStartR(name string, run Run) {
	m.PreStart(NewRun(name, run))
}

func (m *Module) PreStartVR(name string, run voidRun) {
	m.PreStart(newVoidRun(name, run))
}

type PostStarter interface {
	PostStart(hooks ...Runnable)
}

func (m *Module) PostStart(hooks ...Runnable) {
	m.postStart = append(m.postStart, hooks...)
}

func (m *Module) PostStartR(name string, run Run) {
	m.PostStart(NewRun(name, run))
}

func (m *Module) PostStartVR(name string, run voidRun) {
	m.PostStart(newVoidRun(name, run))
}

type PreStopper interface {
	PreStop(hooks ...Runnable)
}

func (m *Module) PreStop(hooks ...Runnable) {
	m.preStop = append(m.preStop, hooks...)
}

func (m *Module) PreStopR(name string, run Run) {
	m.PreStop(NewRun(name, run))
}

func (m *Module) PreStopVR(name string, run voidRun) {
	m.PreStop(newVoidRun(name, run))
}

type PostStopper interface {
	PostStop(hooks ...Runnable)
}

func (m *Module) PostStop(hooks ...Runnable) {
	m.postStop = append(m.postStop, hooks...)
}

func (m *Module) PostStopR(name string, run Run) {
	m.PostStop(NewRun(name, run))
}

func (m *Module) PostStopVR(name string, run voidRun) {
	m.PostStop(newVoidRun(name, run))
}
