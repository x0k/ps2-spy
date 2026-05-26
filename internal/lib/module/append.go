package module

func (m *Module) Go(runnable Runnable) {
	m.services = append(m.services, runnable)
}
