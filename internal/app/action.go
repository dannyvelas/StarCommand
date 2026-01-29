package app

import "fmt"

type action string

const (
	RunAction   action = "run"
	AddAction   action = "add"
	ApplyAction action = "apply"
)

func StringToAction(s string) (action, error) {
	a := action(s)
	switch a {
	case RunAction, AddAction, ApplyAction:
		return a, nil
	default:
		return "", fmt.Errorf("invalid action: %v", s)
	}
}
