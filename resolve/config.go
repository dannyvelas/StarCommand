package resolve

type Config interface {
	Validate() map[string]string
	FillInKeys() error
}
