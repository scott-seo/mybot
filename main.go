package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"strings"

	"time"

	"github.com/peterh/liner"
)

var debug = flag.Bool("debug", true, "debugging")

var historyFn = filepath.Join(os.TempDir(), ".liner_history")

var healthCheck = filepath.Join(os.TempDir(), ".health_check")

func main() {
	flag.Parse()
	// IndexCity()

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
		commandLineCmd := false

		if len(os.Args) > 1 {
			command = strings.Join(os.Args[1:], " ")
			commandLineCmd = true
		} else {
			command, err = line.Prompt("mybot> ")
		}

		if err == nil {
			line.AppendHistory(command)

			tokens := strings.Split(command, " ")
			for _, cmd := range commands {
				if cmd.verb == tokens[0] {
					action := cmd.action

					if len(tokens) > 0 {
						action(strings.Join(tokens[1:], " "))
					} else {
						action("")
					}
					if commandLineCmd {
						time.Sleep(time.Second * 5)
						goto end
					}
					if cmd.verb == "go" {
						time.Sleep(time.Second * 10)
					}
					goto next
				}
			}

			// bashcmd(tokens)

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
