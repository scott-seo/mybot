package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type command struct {
	verb            string
	targets         []string
	action          func([]string)
	secWordComplete func(string) []string
}

var commands []command = []command{}

// this is interesting if the initilization was done at variable assignment
// go complains about initialization loop but
// in init it does not
func init() {
	commands = []command{
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
		command{
			"gmail",
			[]string{},
			gmail,
			nil,
		},
		command{
			"google",
			[]string{},
			google,
			nil,
		},
		command{
			"alert",
			[]string{"warning", "info", "end"},
			alert,
			nil,
		},
		command{
			"graph",
			[]string{"warning", "info", "end"},
			graph,
			nil,
		},
		command{
			"healthcheck",
			[]string{},
			health,
			nil,
		},
		command{
			"repeat",
			[]string{},
			repeat,
			nil,
		},
		command{
			"wait",
			[]string{},
			wait,
			nil,
		},
		command{
			"go",
			[]string{},
			goroutine,
			nil,
		},
	}
}

func google(args []string) {
	q := strings.Join(args, "+")
	bashcmd([]string{"open", "-a", "Google Chrome", fmt.Sprintf("https://www.google.com/#q=%s", q)})
}

func gmail(args []string) {
	bashcmd([]string{"open", "-a", "Google Chrome", "https://mail.google.com"})
}

func hello(args []string) {
	fmt.Println("hello " + strings.Join(args[0:], " "))
}

func alert(args []string) {
	go bashcmd([]string{"afplay", fmt.Sprintf("./alert_%s.mp3", args[0])})
}

func graph(args []string) {
	bashcmd([]string{"open", "-a", "Google Chrome", "./graph.svg"})
}

func health(args []string) {
	resp, err := http.Get(args[0])
	if err != nil {
		fmt.Println(err)
	}

	status := resp.Status

	fmt.Printf("status = %s \n", status)
}

// wait <seconds> <cmd> <args>
func wait(args []string) {
	seconds, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("expecting number but got % instead \n", args[0])
	}

	// execute wait
	// fmt.Printf("wating %d seconds \n", seconds)
	time.Sleep(time.Second * time.Duration(seconds))

	// call chained command
	if len(args) > 2 {
		cmd := findCommand(args[1])

		cmd.action(args[2:])
	}
}

// goroutine <chained commands>
func goroutine(args []string) {
	// call chained command
	if len(args) > 0 {
		cmd := findCommand(args[0])

		go cmd.action(args[1:])
	}
}

// repeat <count> <cmd> <args>
func repeat(args []string) {
	count, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("expecting number but got % instead \n", args[0])
	}

	cmd := findCommand(args[1])

	for i := 0; i < count; i++ {
		cmd.action(args[2:])
	}
}

func findCommand(verb string) command {

	for _, cmd := range commands {
		if cmd.verb == verb {
			return cmd
		}
	}

	return command{}
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
