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
	terraform token = iota
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
		scanner := newScanner(arg)
		_ = scanner.scan()

		// parser := newParser(tokens)
		// parser.parseTargets()
	}
	return nil, nil
}

type scanner struct {
	source string
	tokens []token
}

func newScanner(source string) *scanner {
	return &scanner{
		source: source,
	}
}

func (s *scanner) scan() []token {
	start, current := 0, 0
	for !s.isAtEnd(current) {
		start = current
		s.scanToken(start, current)
	}

	s.tokens = append(s.tokens, eof)
	return s.tokens
}

func (s *scanner) scanToken(start, current int) {
	newCurrent, c := s.advance(current)

	if c == ':' {
		s.tokens = append(s.tokens, colon)
		return
	}

	if s.isLower(c) {
		s.identifier(start, newCurrent)
		return
	}

	fmt.Println("Unexpected character.")
}

func (s *scanner) identifier(start, current int) {
	for s.isLower(s.peek(current)) {
		s.advance(current)
	}

	lexeme := s.source[start:current]
	tok, ok := m[lexeme]
	if !ok {
		fmt.Printf("unrecognized token")
		return
	}

	s.tokens = append(s.tokens, tok)
}

func (s *scanner) isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func (s *scanner) advance(current int) (int, byte) {
	return current + 1, s.source[current]
}

func (s *scanner) peek(current int) byte {
	if s.isAtEnd(current) {
		return 0
	}

	return s.source[current]
}

func (s *scanner) isAtEnd(current int) bool {
	return current >= len(s.source)
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
