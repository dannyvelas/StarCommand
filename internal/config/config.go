package config

type Config interface {
	Validate() map[string]string
	RequiredKeys() []string
	FillInKeys() error
}
