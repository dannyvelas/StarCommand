package config

const (
	statusMissing = "missing"
	statusLoaded  = "loaded"
)

type validateResult struct {
	KeyName string
	Result  string
}

type config interface {
	// Validate returns an map of validation results where each element corresponds to a key in the config
	// the second return value will be false if at least one key was invalid. otherwise, it will be true
	Validate() (map[string]string, bool, error)
}

type fillableConfig interface {
	// FillInKeys takes the keys that are required and uses them to fill out remaining config fields
	FillInKeys() error
}
