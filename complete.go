package main

import (
	"fmt"
	"strings"
)

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
	secondWordPartial := ""
	if len(terms) > 1 {
		secondWordPartial = terms[1]
	}

	fmt.Printf("\n first word = %s \n", firstword)
	fmt.Println("second word = " + secondWordPartial)

	for _, command := range commands {
		if firstword == command.verb {

			if command.secWordComplete != nil && secondWordPartial != "" {
				return prefix, command.secWordComplete(secondWordPartial), ""
			}

			return prefix, command.targets, ""
		}
	}

	return prefix, []string{}, ""
}
