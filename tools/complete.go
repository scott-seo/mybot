package tools

import "strings"

type Command interface {
	Verb() string
	Choices() []string
	ActionFunc() func(string) string
	SecWordCompleteFunc() func(string) []string
}

type Session struct {
	Commands []Command
}

func NewSession(commands []Command) Session {
	return Session{commands}
}

var ExecutionHist []string = make([]string, 0)

func (s Session) WordCompleter(line string, pos int) (head string, completions []string, tail string) {

	prefix := line[0:pos]

	// first word completion
	if !strings.Contains(prefix, " ") {
		var result []string
		for _, command := range s.Commands {
			if strings.HasPrefix(command.Verb(), line) {
				result = append(result, command.Verb())
			}
		}

		// return list of commands
		return "", result, ""
	}

	// second word completion
	terms := strings.Split(line, " ")
	firstword := terms[0]

	for _, command := range s.Commands {
		if firstword == command.Verb() {

			if command.SecWordCompleteFunc() != nil && len(terms) > 1 {
				// mybot>weather__New_York

				return prefix, command.SecWordCompleteFunc()(strings.Join(terms[1:], " ")), ""
			}

			result := []string{}
			sTerm := strings.Join(terms[1:], " ")

			for _, t := range command.Choices() {
				if strings.HasPrefix(t, sTerm) {
					n := strings.Replace(t, sTerm, "", -1)
					result = append(result, n)
				}
			}
			return prefix, result, ""
		}
	}

	return prefix, []string{}, ""
}
