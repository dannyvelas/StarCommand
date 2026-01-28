package main

import (
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

			fmt.Printf("Configs needed:\n%s\n", hostAlias, app.DiagnosticsToTable(diagnostics))
		},
	}

	return checkCmd
}

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

type token int

const (
	invalid token = iota
	terraform
	ssh
	check

	ansible
	inventory
	playbook

	add
	run
	apply

	colon

	eof
)

func toTargets(args []string) ([]app.Target, error) {
	for _, arg := range args {
		_, _ = scan(arg)

		// parser := newParser(tokens)
		// parser.parseTargets()
	}
	return nil, nil
}

func scan(source string) ([]token, []error) {
	errors := make([]error, 0)
	tokens := make([]token, 0)
	start, current := 0, 0
	for current < len(source) {
		start = current
		token, err := scanToken(source, start, current)
		if err != nil {
			errors = append(errors, err)
			continue
		}
		tokens = append(tokens, token)
	}

	tokens = append(tokens, eof)
	return tokens, errors
}

func scanToken(source string, start, current int) (token, error) {
	newCurrent, c := advance(source, current)

	if c == ':' {
		return colon, nil
	}

	if isLower(c) {
		return identifier(source, start, newCurrent)
	}

	return invalid, fmt.Errorf("invalid token")
}

func identifier(source string, start, current int) (token, error) {
	for isLower(peek(source, current)) {
		advance(source, current)
	}

	lexeme := source[start:current]
	tok, ok := m[lexeme]
	if !ok {
		return invalid, fmt.Errorf("unrecognized token")
	}

	return tok, nil
}

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func advance(source string, current int) (int, byte) {
	return current + 1, source[current]
}

func peek(source string, current int) byte {
	if current >= len(source) {
		return 0
	}

	return source[current]
}

//type parser struct {
//	tokens  []token
//	current int
//}
//
//func newParser(tokens []token) *parser {
//	return &parser{
//		tokens:  tokens,
//		current: 0,
//	}
//}

//func (p *parser) parseTargets() []app.Target {
//	targets := make([]app.Target, 0)
//	for !p.isAtEnd() {
//		targets = append(targets, p.parseTarget())
//	}
//	return targets
//}

//func (p *parser) parseTarget() app.Target {
//	if p.match(ansible) {
//	}
//
//	return p.singleResource()
//}
//
//func (p *parser) singleResource() {
//}

//func (p *parser) match(tokens ...token) bool {
//	for _, token := range tokens {
//		if p.check(token) {
//			p.advance()
//			return true
//		}
//	}
//	return false
//}

//func (p *parser) check(token token) bool {
//	if p.isAtEnd() {
//		return false
//	}
//	return p.peek().token == token
//}
//
//func (p *parser) advance() token {
//	if !p.isAtEnd() {
//		p.current += 1
//	}
//	return previous()
//}
//
//func (p *parser) isAtEnd() bool {
//	return p.peek().token == EOF
//}
//
//func (p *parser) peek() token {
//	return p.tokens[p.current]
//}
//
//func (p *parser) previous() token {
//	return p.tokens[p.current-1]
//}
