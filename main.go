package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/j6n/noye/store"

	"github.com/j6n/noye/config"

	"github.com/j6n/noye/ext"
	"github.com/j6n/noye/mock"
	"github.com/j6n/noye/noye"
)

var (
	commands map[string]Command
	sandbox  noye.Manager

	scanner = bufio.NewScanner(os.Stdin)
	output  = make(chan string)

	conf = config.NewConfig()
	db   = store.NewDB()

	opts = map[string]*option{
		"from": &option{"from", "test", "nick of person sending the messages"},
		"chan": &option{"chan", "#noye", "the channel which messages are sent to"},
	}
)

func init() {
	log.SetFlags(0)

	m := conf.ToMap()
	for k, v := range m {
		db.Set("config", k, v)
	}

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
	commands = map[string]Command{
		"l": load(),
		"q": quit(),
		"s": set(),

		"d": dump(),
		"$": debug(),
		"v": source(),

		"#": chanMsg(),
		">": privMsg(),
		".": rawMsg(),

		":": broadcast(true),
		"!": blacklist(),

		"?": help(),
	}
}

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
	for s, cmd := range commands {
		help = append(help, fmt.Sprintf("[%s] %s", s, cmd.Name()))
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
		cmd.Do(line, fields...)
		return
	}

	printHelp()
}
