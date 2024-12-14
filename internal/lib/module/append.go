package module

func (m *Module) Append(services ...Runnable) {
	m.services = append(m.services, services...)
}

func (m *Module) AppendR(name string, run Run) {
	m.Append(NewRun(name, run))
}

func (m *Module) AppendVR(name string, run voidRun) {
	m.Append(newVoidRun(name, run))
}
