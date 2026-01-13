package config

type unvalidatedReader interface {
	ReadUnvalidated() (map[string]string, error)
}

type diagnosticReader interface {
	GetDiagnosticMap() map[string]string
}
