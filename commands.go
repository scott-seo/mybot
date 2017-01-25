package main

import "fmt"

type command struct {
	verb    string
	targets []string
	action  func([]string)
}

var commands []command = []command{
	command{
		"hello",
		[]string{"foo", "bar"},
		hello,
	},
	command{
		"ssh",
		[]string{},
		SSHAction,
	},
}

func hello(args []string) {
	fmt.Println(args[0])
}
