package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"code.google.com/p/go.exp/fsnotify"

	"github.com/j6n/noye/core/config"
	"github.com/j6n/noye/core/script"
	"github.com/j6n/noye/core/store"
	"github.com/j6n/noye/noye"
)

var (
	commands map[string]Command
	sandbox  noye.Manager
	reloader *fsnotify.Watcher

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
	var err error
	if reloader, err = fsnotify.NewWatcher(); err != nil {
		log.Fatalf("creating reloader: %s\n", err)
	}

	log.SetFlags(0)

	mock := noye.NewMockBot()
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

	sandbox = script.New(mock)
	commands = map[string]Command{
		"l": load(),
		"q": quit(),
		"s": set(),

		"d": dump(),
		"$": debug(),
		"v": source(),
		"r": reload(),

		"#": chanMsg(),
		">": privMsg(),
		".": rawMsg(),

		":": broadcast(),
		"!": blacklist(),

		"?": help(),
	}
}

func main() {
	wait := make(chan os.Signal, 1)
	signal.Notify(wait, os.Interrupt)

	go evLoop() // start auto-reload loop

	// TODO find an alternative to \r for terminals that don't support it?
	go func() {
		for {
			fmt.Println("\r<", <-output)
			fmt.Printf("%s> ", "noye")
		}
	}()

	// copy config to db
	for k, v := range conf.ToMap() {
		db.Set("config", k, v)
	}

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

func evLoop() {
	for {
		select {
		case ev := <-reloader.Event:
			switch {
			case ev.IsDelete():
				loadAndWatch(ev.Name)
			}
		case err := <-reloader.Error:
			log.Printf("err watching: %s\n", err)
		}
	}
}
