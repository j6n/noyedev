package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/j6n/noye/ext"
	"github.com/j6n/noye/mock"
	"github.com/j6n/noye/noye"
)

func main() {
	wait := make(chan os.Signal, 1)
	signal.Notify(wait, os.Interrupt)

	// TODO find an alternative to \r for terminals that don't support it?
	go func() {
		for {
			fmt.Println("\r<", <-output)
			fmt.Printf("%s> ", "noye")
		}
	}()

	fmt.Printf("%s> ", "noye")
	for err := scanner.Err(); err == nil && scanner.Scan(); {
		handle(scanner.Text())
		fmt.Printf("%s> ", "noye")
	}

	<-wait
}

func printHelp() {
	var help []string
	for cmd := range commands {
		help = append(help, cmd)
	}
	log.Println("list of commands:", strings.Join(help, ", "))
}

func handle(line string) {
	line = strings.TrimSpace(line)
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return
	}

	if cmd, ok := commands[fields[0]]; ok {
		cmd(line, fields...)
	} else {
		printHelp()
	}
}

// methods for the REPL
func load(line string, parts ...string) {
	if len(parts) == 1 {
		log.Println("usage: 'load path/to/file'")
		return
	}

	file := parts[1]
	if err := sandbox.Load(file); err != nil {
		log.Printf("error loading '%s': %s\n", file, err)
	}
}

func quit(line string, parts ...string) {
	os.Exit(0)
}

func set(line string, parts ...string) {
	so := func(in string) bool {
		return parts[1] == in && len(parts) > 2
	}
	switch {
	case so("ch"):
		log.Printf("setting channel to '%s'\n", parts[1])
		channel = parts[1]
	case so("from"):
		log.Printf("setting from to '%s'\n", parts[1])
		from = parts[1]
	default:
		log.Println("available options: 'ch #channel', 'from user'")
	}
}
func ch(line string, parts ...string) {
	if len(parts) == 1 {
		return
	}

	msg := noye.Message{from, channel, strings.Join(parts[1:], " ")}
	sandbox.Respond(msg)
}

func pm(line string, parts ...string) {
	if len(parts) == 1 {
		return
	}

	msg := noye.Message{from, "noye", strings.Join(parts[1:], " ")}
	sandbox.Respond(msg)
}

func raw(line string, parts ...string) {
	if len(parts) == 1 {
		return
	}
	msg := noye.IrcMessage{
		Source:  from + "!user@localhost",
		Command: parts[1],
	}
	if len(parts) > 2 {
		msg.Args = parts[2:]
	}

	sandbox.Listen(msg)
}

func dump(line string, parts ...string) {
	log.Printf("from: '%s'\n", from)
	log.Printf("channel: '%s'\n", channel)

	log.Println("loaded scripts:")
	for k, v := range sandbox.Scripts() {
		log.Printf("%s @ %s\n", k, v.Path())
	}
}

func source(line string, parts ...string) {
	if len(parts) < 1 {
		return
	}

	for _, script := range sandbox.Scripts() {
		if script.Name() != parts[1] {
			continue
		}

		log.Printf("source for '%s' located at '%s'\n", parts[1], script.Path())
		log.Println(strings.TrimSpace(script.Source()))
	}
}

var (
	commands map[string]command
	sandbox  noye.Manager
	scanner  = bufio.NewScanner(os.Stdin)
	output   = make(chan string)
	from     = "test"
	channel  = "#noye"
)

// TODO make this a struct so commands can be more than just funcs
type command func(string, ...string)

func init() {
	log.SetFlags(0)

	mock := mock.NewMockBot()
	mock.PrivmsgFn = func(target, msg string) {
		output <- fmt.Sprintf("(PRIVMSG) %s: %s", target, msg)
	}
	mock.JoinFn = func(target string) {
		output <- fmt.Sprintf("(JOIN) %s", target)
	}
	mock.PartFn = func(target string) {
		output <- fmt.Sprintf("(PART) %s", target)
	}
	mock.SendFn = func(f string, a ...interface{}) {
		output <- fmt.Sprintf("(SEND) %s", fmt.Sprintf(f, a...))
	}

	sandbox = ext.New(mock)
	commands = map[string]command{
		"load":   load,
		"quit":   quit,
		"set":    set,
		"ch":     ch,
		"pm":     pm,
		"raw":    raw,
		"dump":   dump,
		"source": source,
	}
}
