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
	choices         []string
	action          func(string)
	secWordComplete func(string) []string
}

var commands = []command{}

var memory = make(map[string]map[string]string)

// this is interesting if the initilization was done at variable assignment
// go complains about initialization loop but
// in init it does not
func init() {
	commands = []command{
		command{
			"hello",
			[]string{"foo", "bar", "world"},
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
			[]string{"default"},
			put,
			nil,
		},
		command{
			"get",
			[]string{"default"},
			get,
			nil,
		},
		command{
			"map",
			[]string{},
			remap,
			nil,
		},
		command{
			"debug",
			[]string{},
			setdebug,
			nil,
		},
		command{
			"echo",
			[]string{},
			echo,
			nil,
		},
	}
}

func google(arg string) {
	args := strings.Split(arg, " ")
	q := strings.Join(args, "+")
	bashcmd([]string{"open", "-a", "Google Chrome", fmt.Sprintf("https://www.google.com/#q=%s", q)})
}

func gmail(arg string) {
	bashcmd([]string{"open", "-a", "Google Chrome", "https://mail.google.com"})
}

func hello(arg string) {
	fmt.Println("hello " + arg)
}

func echo(arg string) {
	fmt.Println(arg)
}

// alert [warning|info|end]
func alert(arg string) {
	if len(arg) == 0 {
		return
	}
	args := strings.Split(arg, " ")

	go bashcmd([]string{"afplay", fmt.Sprintf("./alert_%s.mp3", arg)})

	// call chained command
	if len(args) > 1 {
		executeNextIfAny(args[1:])
	}
}

// put <key> <field> <value> cmd...
func put(arg string) {
	if len(arg) < 3 {
		return
	}

	args := strings.Split(arg, " ")

	hasNext := len(args) > 3
	nextTermPost := 3

	key := args[0]
	field := args[1]
	value := args[2]

	if memory[key] == nil {
		memory[key] = make(map[string]string)
	}
	m := memory[key]

	m[field] = value

	if *debug {
		// fmt.Println(memory)
		fmt.Printf("=> put %s %s %s\n", key, field, value)
	}

	// call chained command
	if hasNext {
		executeNextIfAny(args[nextTermPost:])
	}
}

func insert(a []string, x string, i int) []string {
	return append(a[:i], append([]string{x}, a[i:]...)...)
}

func print(args []string) {
	if *debug {
		fmt.Printf("   next = %s \n", args)
	}
}

// get <key> <field> [|]
func get(arg string) {
	args := strings.Split(arg, " ")

	required := 2
	key := args[0]
	field := args[1]
	value := memory[key][field]

	if *debug {
		fmt.Printf("=> get %s %s\n", key, field)
		fmt.Printf("   %s\n", value)
	}

	if len(args) > required && args[2] == "|" {
		args = insert(args, value, required+2)
		print(args[required+1:])
		executeNextIfAny(args[required+1:])
	} else {
		fmt.Println(value)
	}

}

func graph(arg string) {
	bashcmd([]string{"open", "-a", "Google Chrome", "./graph.svg"})
}

func transpose(s string) string {
	if strings.HasPrefix(s, "$") {
		return memory["default"][s[1:]]
	}
	return s
}

// healthcheck <url> | <nextTerm>
//              0    1     2
func healthcheck(arg string) {
	args := strings.Split(arg, " ")

	required := 1
	url := transpose(args[0])
	value := ""

	if *debug {
		fmt.Printf("=> healthcheck %s\n", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	value = strconv.Itoa(resp.StatusCode)

	if *debug {
		fmt.Printf("   %s\n", value)
	}

	if len(args) > required && args[1] == "|" {
		args = insert(args, value, required+2)
		print(args[required+1:])
		executeNextIfAny(args[required+1:])
	} else {
		if !*debug {
			fmt.Println(value)
		}
	}
}

// remap <key> <field> | <nextTerm> <value>
func remap(arg string) {
	args := strings.Split(arg, " ")

	stdout := true
	key := args[0]
	field := args[1]
	value := memory[key][field]
	hasNext := len(args) > 2
	nextTermPos := 2

	if hasNext && args[2] == "|" {
		stdout = false
		nextTermPos = 3
		args = insert(args, value, nextTermPos+1)
	}

	if stdout {
		fmt.Printf("out = %s \n", value)
	}

	if *debug {
		fmt.Println(args)
	}

	// call chained command
	if hasNext {
		executeNextIfAny(args[nextTermPos:])
	}
}

// wait <seconds> | <cmd> <arg>
func wait(arg string) {
	args := strings.Split(arg, " ")
	seconds, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("expecting number but got % instead \n", args[0])
		return
	}

	// execute wait
	if *debug {
		fmt.Printf("=> wait %d\n", seconds)
	}
	time.Sleep(time.Second * time.Duration(seconds))

	// call chained command
	if len(args) > 1 {
		executeNextIfAny(args[1:])
	}
}

func executeNextIfAny(args []string) {
	if len(args) > 0 {
		cmd := findCommand(args[0])
		if cmd != nil {
			cmd.action(strings.Join(args[1:], " "))
		}
	}
}

// goroutine <chained commands>
func goroutine(arg string) {
	args := strings.Split(arg, " ")

	// call chained command
	if len(args) > 0 {
		cmd := findCommand(args[0])
		if cmd != nil {
			go cmd.action(strings.Join(args[1:], " "))
		}
	}
}

// repeat <count> <cmd> <args>
func repeat(arg string) {
	args := strings.Split(arg, " ")
	count, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("expecting number but got %s instead \n", args[0])
	}

	cmd := findCommand(args[1])
	if cmd == nil {
		return
	}

	for i := 0; i < count; i++ {
		cmd.action(strings.Join(args[2:], " "))
	}
}

func findCommand(verb string) *command {

	for _, cmd := range commands {
		if cmd.verb == verb {
			return &cmd
		}
	}

	return nil
}

func setdebug(arg string) {
	if len(arg) == 0 {
		fmt.Println(*debug)
		return
	}

	switch arg {
	case "true":
		*debug = true
	case "false":
		*debug = false
	case "on":
		*debug = true
	case "off":
		*debug = false
	}
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
