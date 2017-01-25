package main

import "strings"

func WordCompleter(line string, pos int) (head string, completions []string, tail string) {

	prefix := line[0:pos]
	// suffix := line[pos:]

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
	firstword := prefix[0 : len(prefix)-1]
	for _, command := range commands {
		if firstword == command.verb {
			return prefix, command.targets, ""
		}
	}

	return prefix, []string{}, ""
}
