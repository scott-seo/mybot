package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"strings"

	"github.com/peterh/liner"
	"github.com/scott-seo/mybot/tools"
)

var debug = flag.Bool("debug", true, "debugging")

var historyFn = filepath.Join(os.TempDir(), ".liner_history")

var healthCheck = filepath.Join(os.TempDir(), ".health_check")

func main() {
	args := os.Args
	lines := []string{}

	flag.Parse()

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)

	s := tools.NewSession(Commands)

	line.SetWordCompleter(s.WordCompleter)
	if f, err := os.Open(historyFn); err == nil {
		line.ReadHistory(f)
		f.Close()
	}
	fileProcessed := false

	for {
	next:
		var command string
		var err error

		if !fileProcessed && strings.Contains(args[0], "mybot") && len(args) > 1 && (strings.HasPrefix(args[1], ".") || strings.HasPrefix(args[1], "/")) {
			file, err := os.Open(args[1])
			if err == nil {
				scanner := bufio.NewScanner(file)
				buf := make([]byte, 0, 64*1024)
				scanner.Buffer(buf, 2*1024*1024)

				for scanner.Scan() {
					s := scanner.Text()
					lines = append(lines, s)
				}
			} else {
				fmt.Println(err)
			}
			file.Close()
		} else {
			command, err = line.Prompt("mybot> ")
			lines = append(lines, command)
		}

		for _, currCmd := range lines {
			if err == nil {
				line.AppendHistory(currCmd)

				tokens := strings.Split(currCmd, " ")

				if "exit" == tokens[0] {
					goto end
				}

				for _, cmd := range Commands {
					if cmd.Verb() == tokens[0] {
						action := cmd.ActionFunc()

						if len(tokens) > 0 {
							action(strings.Join(tokens[1:], " "))
						} else {
							action("")
						}
					}
				}

				// bashcmd(tokens)

				// if len(os.Args) > 1 {
				// 	goto end
				// }

				tools.ExecutionHist = []string{}

			} else if err == liner.ErrPromptAborted {
				log.Print("Aborted")
				goto end
			} else {
				fmt.Println("exit")
				goto end
			}
		}
		lines = []string{}
		fileProcessed = true

		goto next
	}

end:
	if f, err := os.Create(historyFn); err != nil {
		log.Print("Error writing history file: ", err)
	} else {
		line.WriteHistory(f)
		f.Close()
	}

}
