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

var commands = []command{}

var memory = make(map[string]string)

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
			healthcheck,
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
		command{
			"put",
			[]string{},
			put,
			nil,
		},
		command{
			"get",
			[]string{},
			get,
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

// alert [warning|info|end]
func alert(args []string) {
	go bashcmd([]string{"afplay", fmt.Sprintf("./alert_%s.mp3", args[0])})

	// call chained command
	if len(args) > 1 {
		executeNextIfAny(args[1:])
	}
}

// put <key> <value> cmd...
func put(args []string) {

	memory[args[0]] = args[1]
	fmt.Println(memory)

	// call chained command
	if len(args) > 2 {
		executeNextIfAny(args[2:])
	}
}

func insert(a []string, x string, i int) []string {
	return append(a[:i], append([]string{x}, a[i:]...)...)
}

// get <key> cmd...
func get(args []string) {
	var stdout = true
	var key = args[0]
	var value = memory[key]
	var hasNext = len(args) > 1
	var nextTermPost = 1

	if hasNext && args[1] == "|" {
		stdout = false
		nextTermPost = 2
		args = insert(args, value, nextTermPost+1)
	}

	if *debug {
		fmt.Printf("args = %s \n", args)
		fmt.Printf("stdout = %v \n", stdout)
	}

	if stdout {
		fmt.Println(value)
	}

	// call chained command
	if hasNext {
		executeNextIfAny(args[nextTermPost:])
	}
}

func graph(args []string) {
	bashcmd([]string{"open", "-a", "Google Chrome", "./graph.svg"})
}

// healthcheck <url> |
func healthcheck(args []string) {
	var stdout = true
	var url = args[0]
	var value = ""
	var hasNext = len(args) > 1
	var nextTermPost = 1

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	value = fmt.Sprintf("%d", resp.StatusCode)

	if hasNext && args[1] == "|" {
		stdout = false
		nextTermPost = 2
		args = insert(args, value, nextTermPost+1)
	}

	if stdout {
		fmt.Printf("status code = %s \n", value)
	}

	if *debug {
		fmt.Println(args)
	}

	// call chained command
	if len(args) > 1 {
		executeNextIfAny(args[nextTermPost:])
	}
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
	if len(args) > 1 {
		executeNextIfAny(args[1:])
	}
}

func executeNextIfAny(args []string) {
	if len(args) > 0 {
		cmd := findCommand(args[0])
		cmd.action(args[1:])
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
