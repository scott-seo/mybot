package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"os"

	"github.com/scott-seo/mybot/tools"
	"github.com/scott-seo/mybot/weather"
	"golang.org/x/net/websocket"
)

type SimpleCommand struct {
	verb            string
	choices         []string
	action          func(string) string
	secWordComplete func(string) []string
	usage           string
}

func (c SimpleCommand) Verb() string {
	return c.verb
}

func (c SimpleCommand) Choices() []string {
	return c.choices
}

func (c SimpleCommand) ActionFunc() func(string) string {
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
		SimpleCommand{
			"slack",
			[]string{"on", "off"},
			slack,
			nil,
			"",
		},
	}
}

func help(args string) string {
	var output string
	for _, cmd := range Commands {
		c := cmd.(SimpleCommand)
		output += fmt.Sprintf("  %-15s%-15s\n", c.verb, c.usage)
	}
	return output
}

func google(arg string) string {
	args := strings.Split(arg, " ")
	q := strings.Join(args, "+")
	tools.Bashcmd([]string{"open", "-a", "Google Chrome", fmt.Sprintf("https://www.google.com/#q=%s", q)})
	return ""
}

func gmail(arg string) string {
	tools.Bashcmd([]string{"open", "-a", "Google Chrome", "https://mail.google.com"})
	return ""
}

func hello(arg string) string {
	return "hello " + arg
}

func piped(required int, args []string, value string) (bool, string) {
	var output string
	if len(args) > required && args[required] == "|" {
		args = insert(args, value, required+2)
		if *debug {
			output += dformat(strings.Join(args[required+1:], " ")) + "\n"
		}

		output += executeNextIfAny(args[required+1:])
		return false, output
	}
	return false, output
}

func echo(arg string) string {
	var output string

	if strings.Trim(arg, " ") == "" {
		return "\n"
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
		output += dformat(fmt.Sprintf("echo \"%s\"\n", value))
		output += dformat(fmt.Sprintf("%s\n", value))
		// fmt.Printf("   %s\n", args)
	}

	if piped, s := piped(0, args, "\""+value+"\""); !piped {
		output += value + "\n"
		output += s
	}

	return output
}

func say(arg string) string {
	tools.Bashcmd([]string{"say", arg})
	return "say " + arg
}

func blackhole(arg string) string {
	if *debug {
		return dformat("blackhole\n")
	}
	return ""
}

func ifStatement(arg string) string {
	var output string
	args := strings.Split(arg, " ")
	required := 2

	left := args[0]
	right := args[1]

	equals := left == right
	not := strings.HasPrefix(right, "!")
	result := (not && !equals) || equals

	if *debug {
		output += dformat(fmt.Sprintf("if %s == %s is %t\n", left, right, result))
	}

	if len(args) > required && args[2] == "|" {
		if result {
			output += fmt.Sprintf("if %s == %s is %t\n", left, right, result)
			output += executeNextIfAny(args[required+1:])
		}
	}

	return output
}

// alert [warning|info|end]
func alert(arg string) string {
	var output string
	if len(arg) == 0 {
		return ""
	}
	args := strings.Split(arg, " ")

	if *debug {
		output += dformat(fmt.Sprintf("=> alert %s\n", arg))
		output += dformat(fmt.Sprintf("   afplay alert_%s.mp3\n", arg))
	}

	tools.Bashcmd([]string{"afplay", fmt.Sprintf("./alerts/alert_%s.mp3", arg)})

	// call chained command
	if len(args) > 1 {
		output += executeNextIfAny(args[1:])
	}

	return output
}

// put <key> <field> <value> cmd...
func put(arg string) string {
	var output string
	if len(arg) < 3 {
		return ""
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

	output += "put " + fmt.Sprintf("%s %s %s\n", key, field, value)

	if *debug {
		// fmt.Println(memory)
		output += dformat(fmt.Sprintf("put %s %s %s\n", key, field, value))
	}

	// call chained command
	if hasNext {
		output += executeNextIfAny(args[nextTermPost:])
	}

	return output
}

func insert(a []string, x string, i int) []string {
	return append(a[:i], append([]string{x}, a[i:]...)...)
}

// get <key> <field> [|]
func get(arg string) string {
	var output string
	args := strings.Split(arg, " ")

	required := 2
	key := args[0]
	field := args[1]
	value := memory[key][field]

	if *debug {
		output += dformat(fmt.Sprintf("get %s %s\n", key, field))
		output += dformat(fmt.Sprintf("   %s\n", value))
	}

	if piped, s := piped(required, args, value); !piped {
		output += fmt.Sprintln(value)
		output += s
	}

	return output
}

func graph(arg string) string {
	tools.Bashcmd([]string{"open", "-a", "Google Chrome", "./asset/graph.svg"})
	return ""
}

func transpose(s string) string {
	if strings.HasPrefix(s, "$") {
		return memory["default"][s[1:]]
	}
	return s
}

// healthcheck <url> | <nextTerm>
//              0    1     2
func healthcheck(arg string) string {
	var output string
	args := strings.Split(arg, " ")

	required := 1
	url := transpose(args[0])
	value := ""

	resp, err := http.Get(url)
	if err != nil {
		output += eformat(fmt.Sprintln(err))
	}

	value = strconv.Itoa(resp.StatusCode)

	if piped, s := piped(required, args, value); !piped {
		if *debug {
			output += dformat(fmt.Sprintf("healthcheck %s | %s\n", url, s))
		}
		output += fmt.Sprintf("healthcheck status code is %s\n", value)
	}

	return output
}

// wait <seconds> | <cmd> <arg>
func wait(arg string) string {
	var output string
	args := strings.Split(arg, " ")
	seconds, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Sprintf("expecting number but got %v instead \n", args[0])
	}

	// execute wait
	if *debug {
		output += dformat(fmt.Sprintf("wait %d\n", seconds))
	}
	time.Sleep(time.Second * time.Duration(seconds))

	// call chained command
	if len(args) > 1 {
		output += executeNextIfAny(args[1:])
	}

	return output
}

func executeNextIfAny(args []string) string {
	if len(args) > 0 {
		cmd := findCommand(args[0])
		if cmd != nil {
			return cmd.action(strings.Join(args[1:], " "))
		}
	}
	return ""
}

func monitor(arg string) string {
	var output string
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
		return ""
	case "add":
		interval, _ := strconv.Atoi(args[1])
		monitorID++
		monitors[monitorID] = true
		monitorsDetail[monitorID] = strings.Join(args[2:], " ")
		registerMonitor(monitorID, interval, args[2:])
		return ""
	case "ls":
		for k, v := range monitorsDetail {
			output += fmt.Sprintf("[%d] %s\n", k, v)
		}
		return output
	}

	return ""
}

func registerMonitor(id int, interval int, args []string) {

	go func() {
		for {
			if *debug {
				fmt.Printf(dformat(fmt.Sprintf("running [%d] %s \n", id, monitorsDetail[id])))
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
func repeat(arg string) string {
	var output string
	args := strings.Split(arg, " ")
	count, err := strconv.Atoi(args[0])
	if err != nil {
		output += dformat(fmt.Sprintf("expecting number but got %s instead \n", args[0]))
	}

	cmd := findCommand(args[1])
	if cmd == nil {
		return output
	}

	for i := 0; i < count; i++ {
		output += dformat(cmd.action(strings.Join(args[2:], " ")))
	}

	return output
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

func dformat(s string) string {
	return "\u001B[0;32mdebug> " + s + "\u001B[0m"
}

func eformat(s string) string {
	return "\u001B[0;31merror> " + s + "\u001B[0m"
}

func setdebug(arg string) string {
	var output string
	output += dformat("debug " + arg + "\n")

	switch arg {
	case "true":
		*debug = true
		output += "debug on\n"
	case "false":
		*debug = false
		output += "debug off\n"
	case "on":
		*debug = true
		output += "debug on\n"
	case "off":
		*debug = false
		output += "debug off\n"
	default:
		value := "off"
		if *debug {
			value = "on"
		}
		return "debug " + value + "\n"
	}

	return output
}

var ws *websocket.Conn
var id string

func slack(arg string) string {
	var output string
	if arg == "on" && ws == nil {
		ws, id = tools.SlackConnect(os.Getenv("SLACK_BOT_TOKEN"))

		go func() {
			for {
				// read each incoming message
				m, err := tools.GetMessage(ws)
				if err != nil {
					log.Fatal(err)
				}

				// d, _ := json.MarshalIndent(m, "", " ")
				// fmt.Println(string(d))

				// see if we're mentioned
				if m.Type == "message" && strings.HasPrefix(m.Text, "<@"+id+">") {

					parts := strings.Fields(m.Text)
					output += fmt.Sprintf("\nuser : %s\n", strings.Join(parts[1:], " "))
					if len(parts) > 1 {
						executeNextIfAny(parts[1:])
					}

					m.Text = strings.Join(tools.ExecutionHist, "\n")
					// fmt.Printf("mybot: %s\n", m.Text)
					tools.PostMessage(ws, m)
				}
			}
		}()
	}

	if arg == "off" && ws != nil {
		go func() {
			ws.Close()
		}()
	}

	return output
}
