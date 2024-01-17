package multi_loaders

const DefaultLoader = "default"

func LoaderName(loader string) string {
	if loader == "" {
		return DefaultLoader
	}
	return loader
}

type MultiLoader interface {
	Loaders() []string
}
