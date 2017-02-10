package tools

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
)

// Bashcmd will execute binary
func Bashcmd(args []string) {
	_, err := exec.LookPath(args[0])
	if err != nil {
		fmt.Printf("command [%s] not found \n", args[0])
		return
	}

	cmd := exec.Command(args[0], args[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}
