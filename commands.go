package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/scott-seo/mybot/tools"
	"github.com/scott-seo/mybot/weather"
)

type SimpleCommand struct {
	verb            string
	choices         []string
	action          func(string)
	secWordComplete func(string) []string
	usage           string
}

func (c SimpleCommand) Verb() string {
	return c.verb
}

func (c SimpleCommand) Choices() []string {
	return c.choices
}

func (c SimpleCommand) ActionFunc() func(string) {
	return c.action
}

func (c SimpleCommand) SecWordCompleteFunc() func(string) []string {
	return c.secWordComplete
}

var Commands = []tools.Command{}

var memory = make(map[string]map[string]string)

var monitors = make(map[int]bool)
var monitorsDetail = make(map[int]string)

var monitorID int

// this is interesting if the initilization was done at variable assignment
// go complains about initialization loop but
// in init it does not
func init() {
	Commands = []tools.Command{
		SimpleCommand{
			"hello",
			[]string{"foo", "bar", "world"},
			hello,
			nil,
			"[foo|bar|world]",
		},
		SimpleCommand{
			"ssh",
			[]string{},
			tools.SSHAction,
			nil,
			"[user] [host] [\"command\"]",
		},
		SimpleCommand{
			"weather",
			[]string{},
			weather.Action,
			weather.CitySearch,
			"[City Name]",
		},
		SimpleCommand{
			"gmail",
			[]string{},
			gmail,
			nil,
			"",
		},
		SimpleCommand{
			"google",
			[]string{},
			google,
			nil,
			"[search term]",
		},
		SimpleCommand{
			"alert",
			[]string{"warning", "info", "end"},
			alert,
			nil,
			"[warning | info | end]",
		},
		SimpleCommand{
			"graph",
			[]string{"warning", "info", "end"},
			graph,
			nil,
			"",
		},
		SimpleCommand{
			"healthcheck",
			[]string{},
			healthcheck,
			nil,
			"[url]",
		},
		SimpleCommand{
			"repeat",
			[]string{},
			repeat,
			nil,
			"[number of execution]",
		},
		SimpleCommand{
			"wait",
			[]string{},
			wait,
			nil,
			"[time in seconds]",
		},
		SimpleCommand{
			"put",
			[]string{"default"},
			put,
			nil,
			"[key] [field] [value]",
		},
		SimpleCommand{
			"get",
			[]string{"default"},
			get,
			nil,
			"[key] [field]",
		},
		SimpleCommand{
			"debug",
			[]string{},
			setdebug,
			nil,
			"[on|off]",
		},
		SimpleCommand{
			"echo",
			[]string{},
			echo,
			nil,
			"[message wrapped in double quotes]",
		},
		SimpleCommand{
			"monitor",
			[]string{"add", "remove", "ls"},
			monitor,
			nil,
			"[add | remove | ls]",
		},
		SimpleCommand{
			"blackhole",
			[]string{},
			blackhole,
			nil,
			"",
		},
		SimpleCommand{
			"if",
			[]string{},
			ifStatement,
			nil,
			"[value to compare with redirected input]",
		},
		SimpleCommand{
			"say",
			[]string{},
			say,
			nil,
			"[message]",
		},
		SimpleCommand{
			"help",
			[]string{},
			help,
			nil,
			"",
		},
	}
}

func help(args string) {
	for _, cmd := range Commands {
		c := cmd.(SimpleCommand)
		fmt.Printf("  %-15s%-15s\n", c.verb, c.usage)
	}
}

func google(arg string) {
	args := strings.Split(arg, " ")
	q := strings.Join(args, "+")
	tools.Bashcmd([]string{"open", "-a", "Google Chrome", fmt.Sprintf("https://www.google.com/#q=%s", q)})
}

func gmail(arg string) {
	tools.Bashcmd([]string{"open", "-a", "Google Chrome", "https://mail.google.com"})
}

func hello(arg string) {
	fmt.Println("hello " + arg)
}

func piped(required int, args []string, value string) bool {
	if len(args) > required && args[required] == "|" {
		args = insert(args, value, required+2)
		print(args[required+1:])
		executeNextIfAny(args[required+1:])
		return true
	}
	return false
}

func echo(arg string) {
	if strings.Trim(arg, " ") == "" {
		fmt.Println()
		return
	}

	if !strings.Contains(arg, "\"") {
		pos := strings.Index(arg, "|")
		if pos > 0 {
			arg = `"` + arg[0:pos-1] + `"` + arg[pos:]
		} else {
			arg = `"` + arg + `"`
		}

	}

	args := []string{}

	endPos := strings.Index(arg, "|")

	if endPos == -1 {
		endPos = len(arg)
	}

	firstQ := strings.Index(arg[0:endPos], `"`)
	secondQ := strings.LastIndex(arg[0:endPos], `"`)
	value := arg[firstQ+1 : secondQ]
	args = strings.Split(strings.Trim(arg[secondQ+1:], " "), " ")

	if *debug {
		fmt.Printf("=> echo \"%s\"\n", value)
		fmt.Printf("   %s\n", value)
		// fmt.Printf("   %s\n", args)
	}

	if piped := piped(0, args, "\""+value+"\""); !piped {
		fmt.Println(value)
	}

}

func say(arg string) {
	tools.Bashcmd([]string{"say", arg})
}

func blackhole(arg string) {
	if *debug {
		fmt.Println("=> blackhole")
	}
}

func ifStatement(arg string) {
	args := strings.Split(arg, " ")
	required := 2

	left := args[0]
	right := args[1]

	equals := left == right
	not := strings.HasPrefix(right, "!")
	result := (not && !equals) || equals

	if *debug {
		fmt.Printf("=> if %s == %s\n", left, right)
		fmt.Printf("   %t\n", result)
	}

	if len(args) > required && args[2] == "|" {
		if result {
			print(args[required+1:])
			executeNextIfAny(args[required+1:])
		}
	}

}

// alert [warning|info|end]
func alert(arg string) {
	if len(arg) == 0 {
		return
	}
	args := strings.Split(arg, " ")

	if *debug {
		fmt.Printf("=> alert %s\n", arg)
		fmt.Printf("   afplay alert_%s.mp3\n", arg)
	}

	tools.Bashcmd([]string{"afplay", fmt.Sprintf("./alerts/alert_%s.mp3", arg)})

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
		fmt.Printf("   next=> %s \n\n", args)
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

	if piped := piped(required, args, value); !piped {
		fmt.Println(value)
	}

}

func graph(arg string) {
	tools.Bashcmd([]string{"open", "-a", "Google Chrome", "./asset/graph.svg"})
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

	if piped := piped(required, args, value); !piped {
		fmt.Println(value)
	}
}

// wait <seconds> | <cmd> <arg>
func wait(arg string) {
	args := strings.Split(arg, " ")
	seconds, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Printf("expecting number but got %v instead \n", args[0])
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

func monitor(arg string) {
	args := strings.Split(arg, " ")

	// monitor off 1
	switch args[0] {
	case "remove":
		id, _ := strconv.Atoi(args[1])
		for k := range monitors {
			if id == k {
				monitors[k] = false
				delete(monitors, k)
				delete(monitorsDetail, k)
			}
		}
		return
	case "add":
		interval, _ := strconv.Atoi(args[1])
		monitorID++
		monitors[monitorID] = true
		monitorsDetail[monitorID] = strings.Join(args[2:], " ")
		registerMonitor(monitorID, interval, args[2:])
		return
	case "ls":
		for k, v := range monitorsDetail {
			fmt.Printf("[%d] %s\n", k, v)
		}
		return
	}

}

func registerMonitor(id int, interval int, args []string) {

	go func() {
		for {
			if *debug {
				fmt.Printf("\n[%d] %s \n", id, monitorsDetail[id])
			}
			if !monitors[id] {
				return
			}

			if len(args) > 1 {
				executeNextIfAny(args)
			}

			time.Sleep(time.Second * time.Duration(interval))
		}
	}()
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

func findCommand(verb string) *SimpleCommand {

	for _, cmd := range Commands {
		if cmd.Verb() == verb {
			c := cmd.(SimpleCommand)
			return &c
		}
	}

	return nil
}

func setdebug(arg string) {
	switch arg {
	case "true":
		*debug = true
	case "false":
		*debug = false
	case "on":
		*debug = true
	case "off":
		*debug = false
	default:
		value := "off"
		if *debug {
			value = "on"
		}
		fmt.Println(value)
	}
}
