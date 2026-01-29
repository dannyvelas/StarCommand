package app

import (
	"fmt"
	"strings"
)

type Target struct {
	resource resource
	action   action
}

func ToTargets(args []string) ([]Target, error) {
	targets := make([]Target, 0)
	for _, arg := range args {
		target, err := toTarget(arg)
		if err != nil {
			return nil, err
		}
		targets = append(targets, target)
	}
	return targets, nil
}

func toTarget(arg string) (Target, error) {
	split := strings.Split(arg, ":")

	resource, rest, err := parseResource(arg, split)
	if err != nil {
		return Target{}, err
	}

	action, err := parseAction(resource, rest)
	if err != nil {
		return Target{}, err
	}

	return Target{resource: resource, action: action}, nil
}

func parseResource(arg string, split []string) (resource, []string, error) {
	first, rest, err := shift(split)
	if err != nil {
		return "", rest, fmt.Errorf("error: invalid target argument: %s", arg)
	}

	switch first {
	case "ansible":
		return parseAnsibleResource(rest)
	case "ssh":
		return SSHResource, rest, nil
	case "terraform":
		return TerraformResource, rest, nil
	default:
		return "", rest, fmt.Errorf("error: unrecognized resource: %s", first)
	}
}

func parseAnsibleResource(split []string) (resource, []string, error) {
	first, rest, err := shift(split)
	if err != nil {
		return "", rest, fmt.Errorf("error: expecting ansible sub-command")
	}

	switch first {
	case "playbook":
		return AnsiblePlaybookResource, rest, nil
	case "inventory":
		return AnsibleInventoryResource, rest, nil
	default:
		return "", rest, fmt.Errorf("error: unrecognized ansible sub-command: %s", first)
	}
}

func parseAction(resource resource, split []string) (action, error) {
	first, _, err := shift(split)
	if err != nil {
		return "", fmt.Errorf("error: expecting action after %s resource", resource)
	}

	action, err := StringToAction(first)
	if err != nil {
		return "", fmt.Errorf("error: unrecognized action (%s) for %s resource", first, resource)
	}

	return action, nil
}

func shift(s []string) (string, []string, error) {
	if len(s) < 1 {
		return "", s, fmt.Errorf("empty slice")
	}

	return s[0], s[1:], nil
}
