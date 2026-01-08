package config

type provider interface {
	Decode(target any) error
}
