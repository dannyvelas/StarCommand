package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dannyvelas/conflux"
	"github.com/dannyvelas/homelab/internal/app"
	"github.com/dannyvelas/homelab/internal/helpers"
	"github.com/spf13/cobra"
)

func checkCmd() *cobra.Command {
	checkCmd := &cobra.Command{
		Use:   "check <host-alias> target1 [targets]",
		Short: "Print a diagnostic report of all the configs that were found/missing for a given resource",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			hostAlias := args[0]
			configMux := conflux.NewConfigMux(
				conflux.WithYAMLFileReader(helpers.FallbackFile, conflux.WithPath(helpers.GetConfigPath(hostAlias))),
				conflux.WithEnvReader(),
				conflux.WithBitwardenSecretReader(),
			)

			targets, err := toTargets(args[1:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			diagnostics, err := app.Check(configMux, hostAlias, targets)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err.Error())
				os.Exit(1)
			}

			fmt.Printf("Configs needed for host(%s):\n%s\n", hostAlias, app.DiagnosticsToTable(diagnostics))
		},
	}

	return checkCmd
}

var errSkip = fmt.Errorf("skip token")

type token string

const (
	nilToken token = ""

	terraform token = "terraform"
	ssh       token = "ssh"
	check     token = "check"

	ansible   token = "ansible"
	inventory token = "inventory"
	playbook  token = "playbook"

	add   token = "add"
	run   token = "run"
	apply token = "apply"

	eof token = "eof"
)

var m = map[string]token{
	"terraform": terraform,
	"ssh":       ssh,
	"check":     check,

	"ansible":   ansible,
	"inventory": inventory,
	"playbook":  playbook,

	"add":   add,
	"run":   run,
	"apply": apply,
}

type tokenctx struct {
	token token
	err   error
}

func toTargets(args []string) ([]app.Target, error) {
	for _, arg := range args {
		tokens := scan(arg)
		target, err := parse(tokens)
		fmt.Println(target, err)
	}
	return nil, errors.New("unimplemented")
}

func scan(source string) chan tokenctx {
	tokens := make(chan tokenctx)
	go func() {
		start, current := 0, 0
		for current < len(source) {
			start = current
			newCurrent, token, err := scanToken(source, start, current)
			current = newCurrent
			if errors.Is(err, errSkip) {
				continue
			} else if err != nil {
				tokens <- tokenctx{token: nilToken, err: err}
				continue
			}
			tokens <- tokenctx{token: token}
		}

		tokens <- tokenctx{token: eof}
	}()
	return tokens
}

func scanToken(source string, start, current int) (int, token, error) {
	newCurrent, c := current+1, source[current]

	if c == ':' {
		return newCurrent, nilToken, errSkip
	}

	if !isLower(c) {
		return newCurrent, nilToken, fmt.Errorf("unexpected character: %v", c)
	}

	for newCurrent < len(source) && isLower(source[newCurrent]) {
		newCurrent += 1
	}

	lexeme := source[start:newCurrent]
	tok, ok := m[lexeme]
	if !ok {
		return newCurrent, nilToken, fmt.Errorf("unexpected target: %v", lexeme)
	}

	return newCurrent, tok, nil
}

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func parse(tokens chan tokenctx) (app.Target, error) {
	resource, err := parseResource(tokens)
	if err != nil {
		return app.Target{}, err
	}

	action, err := parseAction(tokens, resource)
	if err != nil {
		return app.Target{}, err
	}

	tokenctx, ok := <-tokens
	if tokenctx.err != nil {
		return app.Target{}, tokenctx.err
	}

	if ok && tokenctx.token != eof {
		return app.Target{}, fmt.Errorf("unexpected value after \"%s\": %s", action, tokenctx.token)
	}

	return app.Target{Resource: resource, Action: action}, nil
}

func parseResource(tokens chan tokenctx) (app.Resource, error) {
	tokenctx := <-tokens
	if tokenctx.err != nil {
		return "", tokenctx.err
	}

	switch tokenctx.token {
	case ansible:
		return parseAnsibleResource(tokens)
	case ssh:
		return app.SSHResource, nil
	case terraform:
		return app.TerraformResource, nil
	default:
		return "", fmt.Errorf("invalid resource")
	}
}

func parseAnsibleResource(tokens chan tokenctx) (app.Resource, error) {
	tokenctx := <-tokens
	if tokenctx.err != nil {
		return "", tokenctx.err
	}

	switch tokenctx.token {
	case playbook:
		return app.AnsiblePlaybookResource, nil
	case inventory:
		return app.AnsibleInventoryResource, nil
	case eof:
		return "", fmt.Errorf("unexpected end of input. expecting sub-command for \"ansible\"")
	default:
		return "", fmt.Errorf("invalid sub-command for \"ansible\": %v", tokenctx.token)
	}
}

func parseAction(tokens chan tokenctx, resource app.Resource) (app.Action, error) {
	tokenctx := <-tokens
	if tokenctx.err != nil {
		return "", tokenctx.err
	}

	switch tokenctx.token {
	case run:
		return app.RunAction, nil
	case add:
		return app.AddAction, nil
	case apply:
		return app.ApplyAction, nil
	case eof:
		return "", fmt.Errorf("unexpected end of input. expecting an action after \"%s\"", resource)
	default:
		return "", fmt.Errorf("invalid action: %s", tokenctx.token)
	}
}
