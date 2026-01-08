package config

type provider interface {
	UnmarshalInto(target any) error
}

type validatedReader interface {
	ReadValidated() (map[string]string, error)
}

type unvalidatedReader interface {
	ReadUnvalidated() (map[string]string, error)
}
