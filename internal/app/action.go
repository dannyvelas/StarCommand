package app

import "fmt"

type action string

const (
	runAction   action = "run"
	addAction   action = "add"
	applyAction action = "apply"
)

func stringToAction(s string) (action, error) {
	a := action(s)
	switch a {
	case runAction, addAction, applyAction:
		return a, nil
	default:
		return "", fmt.Errorf("invalid action: %v", s)
	}
}
