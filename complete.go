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

	// fmt.Printf("\n terms = %s \n", terms)

	for _, command := range commands {
		if firstword == command.verb {

			if command.secWordComplete != nil && len(terms) > 1 {
				return prefix, command.secWordComplete(terms[1:]), ""
			}

			return prefix, command.targets, ""
		}
	}

	return prefix, []string{}, ""
}
