package app

import (
	"errors"
	"fmt"
	"strings"
)

type target struct {
	resource resource
	action   action
}

var (
	errExpectedEOF                   = errors.New("error: expected EOF")
	errInvalidTargetArgument         = errors.New("error: invalid target argument")
	errUnrecognizedResource          = errors.New("error: unrecognized resource")
	errExpectingAnsibleSubCommand    = errors.New("error: expecting ansible sub-command")
	errUnrecognizedAnsibleSubCommand = errors.New("error: unrecognized ansible sub-command")
	errExpectingAction               = errors.New("error: expecting action")
	errUnrecognizedAction            = errors.New("error: unrecognized action")
	errEmptySlice                    = errors.New("error: empty slice")
)

func toTargets(args []string) ([]target, error) {
	targets := make([]target, 0)
	for _, arg := range args {
		target, err := toTarget(arg)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func toTarget(arg string) (target, error) {
	split := strings.Split(arg, ":")

	resource, rest, err := parseResource(arg, split)
	if err != nil {
		return target{}, err
	}

	action, rest, err := parseAction(resource, rest)
	if err != nil {
		return target{}, err
	}

	if len(rest) != 0 {
		return target{}, fmt.Errorf("%w: saw %v", errExpectedEOF, rest)
	}

	return target{resource: resource, action: action}, nil
}

func parseResource(arg string, split []string) (resource, []string, error) {
	first, rest, err := shift(split)
	if err != nil {
		return "", rest, fmt.Errorf("%w: %s", errInvalidTargetArgument, arg)
	}

	switch first {
	case "ansible":
		return parseAnsibleResource(rest)
	case "ssh":
		return sshResource, rest, nil
	case "terraform":
		return terraformResource, rest, nil
	default:
		return "", rest, fmt.Errorf("%w: %s", errUnrecognizedResource, first)
	}
}

func parseAnsibleResource(split []string) (resource, []string, error) {
	first, rest, err := shift(split)
	if err != nil {
		return "", rest, errExpectingAnsibleSubCommand
	}

	switch first {
	case "playbook":
		return ansiblePlaybookResource, rest, nil
	case "inventory":
		return ansibleInventoryResource, rest, nil
	default:
		return "", rest, fmt.Errorf("%w: %s", errUnrecognizedAnsibleSubCommand, first)
	}
}

func parseAction(resource resource, split []string) (action, []string, error) {
	first, rest, err := shift(split)
	if err != nil {
		return "", rest, fmt.Errorf("%w after %s resource", errExpectingAction, resource)
	}

	action, err := stringToAction(first)
	if err != nil {
		return "", rest, fmt.Errorf("%w (%s) after %s resource", errUnrecognizedAction, first, resource)
	}

	return action, rest, nil
}

func shift(s []string) (string, []string, error) {
	if len(s) < 1 {
		return "", s, errEmptySlice
	}

	return s[0], s[1:], nil
}
