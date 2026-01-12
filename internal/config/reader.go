package config

type unvalidatedReader interface {
	ReadUnvalidated() (map[string]string, error)
}
