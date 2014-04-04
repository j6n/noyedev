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
	for s, name := range commands {
		help = append(help, fmt.Sprintf("[%s] %s", s, name.Name()))
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

// methods for the REPL
func load() Command {
	cmd := newCommand("load path/to/file.js", "load")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			log.Println(cmd.Help())
			return
		}

		file := parts[1]
		if err := sandbox.Load(file); err != nil {
			log.Printf("error loading '%s': %s\n", file, err)
		}
	}
	return cmd
}

func quit() Command {
	cmd := newCommand("quits the REPL", "quit")
	cmd.fn = func(line string, parts ...string) {
		os.Exit(0)
	}
	return cmd
}

func set() Command {
	msg := fmt.Sprintf(`available options: 
  ch: sets the channel
  from: sets the nick messages come from
current:
  ch: %s, from: %s`, channel, from)

	cmd := newCommand("sets REPL options", "set")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			log.Println(msg)
			return
		}

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
			log.Println(msg)
		}
	}
	return cmd
}

func chanMsg() Command {
	cmd := newCommand("send text as a chan msg", "chanmsg")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			log.Println(cmd.Help())
			return
		}

		msg := noye.Message{from, channel, strings.Join(parts[1:], " ")}
		sandbox.Respond(msg)
	}
	return cmd
}

func privMsg() Command {
	cmd := newCommand("send text as a priv msg", "privmsg")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			log.Println(cmd.Help())
			return
		}

		msg := noye.Message{from, "noye", strings.Join(parts[1:], " ")}
		sandbox.Respond(msg)
	}
	return cmd
}

func rawMsg() Command {
	cmd := newCommand("send text as a raw msg", "rawmsg")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			log.Println(cmd.Help())
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
	return cmd
}

func dump() Command {
	cmd := newCommand("dumps the state of the REPL", "dump")
	cmd.fn = func(line string, parts ...string) {
		log.Printf("from: '%s'\n", from)
		log.Printf("channel: '%s'\n", channel)

		log.Println("loaded scripts:")
		for k, v := range sandbox.Scripts() {
			log.Printf("%s @ %s\n", k, v.Path())
		}
	}
	return cmd
}

func source() Command {
	cmd := newCommand("dumps source for script", "source")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			log.Println(cmd.Help())
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
	return cmd
}

func help() Command {
	cmd := newCommand("display help for commands", "help")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			printHelp()
			return
		}

		if cmd, ok := commands[parts[1]]; ok {
			log.Printf("[%s] %s: %s\n", parts[1], cmd.Name(), cmd.Help())
		}
	}
	return cmd
}

func broadcast(private bool) Command {
	cmd := newCommand("brodcasts via the message system", "broadcast")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 || len(parts) < 3 {
			log.Println(parts[0], "<key> <val>")
			return
		}

		ext.Broadcast(parts[1], strings.Join(parts[2:], " "))
	}
	return cmd
}

func blacklist() Command {
	cmd := newCommand("blacklists a key", "blacklist")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 || len(parts) < 2 {
			log.Println(parts[0], "<key1> <key2> ...")
			return
		}

		ext.AddPrivate(parts[1:]...)
	}
	return cmd
}

var (
	commands map[string]Command
	sandbox  noye.Manager
	scanner  = bufio.NewScanner(os.Stdin)
	output   = make(chan string)
	from     = "test"
	channel  = "#noye"
)

type Command interface {
	Help() string
	Name() string
	Do(string, ...string)
}

type command struct {
	help string
	name string
	fn   func(string, ...string)
}

func (c command) Help() string {
	return c.help
}

func (c command) Name() string {
	return c.name
}

func (c command) Do(line string, parts ...string) {
	c.fn(line, parts...)
}

func newCommand(help, name string) command {
	return command{help: help, name: name, fn: func(string, ...string) {}}
}

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
	commands = map[string]Command{
		"l": load(),
		"q": quit(),
		"s": set(),

		"d": dump(),
		"v": source(),

		"#": chanMsg(),
		">": privMsg(),
		".": rawMsg(),

		":": broadcast(true),
		";": broadcast(false),
		"!": blacklist(),

		"?": help(),
	}
}
