package main

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
)

type command struct {
	verb            string
	targets         []string
	action          func([]string)
	secWordComplete func(string) []string
}

var commands []command = []command{
	command{
		"hello",
		[]string{"foo", "bar"},
		hello,
		nil,
	},
	command{
		"ssh",
		[]string{},
		SSHAction,
		nil,
	},
	command{
		"weather",
		[]string{},
		WeatherAction,
		CitySearch,
	},
}

func hello(args []string) {
	fmt.Println(args[0])
}

func bashcmd(args []string) {
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
