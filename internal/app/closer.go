package app

type NamedCloser struct {
	Name  string
	Close func() error
}
