package app

import "fmt"

type Action string

const (
	RunAction   Action = "run"
	AddAction   Action = "add"
	ApplyAction Action = "apply"
)

func StringToAction(s string) (Action, error) {
	a := Action(s)
	switch a {
	case RunAction, AddAction, ApplyAction:
		return a, nil
	default:
		return "", fmt.Errorf("invalid action: %v", s)
	}
}
