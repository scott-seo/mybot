package main

import "strings"

func WordCompleter(line string, pos int) (head string, completions []string, tail string) {

	prefix := line[0:pos]

	// first word completion
	if !strings.Contains(prefix, " ") {
		var result []string
		for _, command := range commands {
			if strings.HasPrefix(command.verb, line) {
				result = append(result, command.verb)
			}
		}

		// return list of commands
		return "", result, ""
	}

	// second word completion
	terms := strings.Split(line, " ")
	firstword := terms[0]

	for _, command := range commands {
		if firstword == command.verb {

			if command.secWordComplete != nil && len(terms) > 1 {
				// mybot>weather__New_York

				return prefix, command.secWordComplete(strings.Join(terms[1:], " ")), ""
			}

			result := []string{}
			sTerm := strings.Join(terms[1:], " ")

			for _, t := range command.choices {
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
