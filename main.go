package main

import (
	"log"
	"os"
	"path/filepath"

	"strings"

	"github.com/peterh/liner"
)

var historyFn = filepath.Join(os.TempDir(), ".liner_history")

func main() {
	IndexCity()

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)
	line.SetWordCompleter(WordCompleter)
	if f, err := os.Open(historyFn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}

	for {
	next:
		var command string
		var err error

		if len(os.Args) > 1 {
			command = strings.Join(os.Args[1:], " ")
		} else {
			command, err = line.Prompt("mybot> ")
		}

		if err == nil {
			line.AppendHistory(command)

			tokens := strings.Split(command, " ")
			for _, cmd := range commands {
				if cmd.verb == tokens[0] {
					f := cmd.action
					if len(tokens) > 0 {
						f(tokens[1:])
					} else {
						f([]string{})
					}
					goto next
				}
			}

			bashcmd(tokens)

			if len(os.Args) > 1 {
				goto end
			}

		} else if err == liner.ErrPromptAborted {
			log.Print("Aborted")
			goto end
		} else {
			// log.Print("Error reading line: ", err)
			goto end
		}
	}

end:
	if f, err := os.Create(historyFn); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}

}
