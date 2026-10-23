package module

// Go registers a critical service. An error returned by a critical service
// cancels the module and is returned from Run.
func (m *Module) Go(runnable Runnable) {
	m.services = append(m.services, service{
		Runnable: runnable,
		critical: true,
	})
}

// GoFeature registers a feature service. An error returned by a feature service
// is logged but does not stop the module.
func (m *Module) GoFeature(runnable Runnable) {
	m.services = append(m.services, service{
		Runnable: runnable,
	})
}
