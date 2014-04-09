package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/j6n/noye/ext"
	"github.com/j6n/noye/irc"
	"github.com/j6n/noye/noye"
	"github.com/j6n/noye/store"
)

func loadAndWatch(file string) {
	if err := sandbox.Load(file); err != nil {
		log.Printf("error (auto)loading '%s': %s\n", file, err)
		return
	}
	if err := reloader.Watch(file); err != nil {
		log.Printf("err watching: %s, %s\n", file, err)
	}
}

// methods for the REPL
func load() Command {
	cmd := newCommand("load path/to/file.js", "load")
	cmd.fn = func(line string, parts ...string) {
		if len(parts) == 1 {
			log.Println(cmd.Help())
			return
		}

		loadAndWatch(parts[1])
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
	cmd := newCommand("sets REPL options", "set")
	cmd.fn = func(line string, parts ...string) {
		msg := getOps()
		if len(parts) < 2 {
			log.Println(msg)
			return
		}

		if o, ok := opts[parts[1]]; ok {
			old, val := o.Val, strings.Join(parts[2:], " ")
			o.Set(val)
			log.Printf("set '%s' to '%s' (was: '%s')\n", o.Name, val, old)
		} else {
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

		msg := noye.Message{
			From:   getUser(),
			Target: opts["chan"].String(),
			Text:   strings.Join(parts[1:], " ")}
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

		msg := noye.Message{
			From:   getUser(),
			Target: "noye",
			Text:   strings.Join(parts[1:], " "),
		}
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
			Source:  getUser(),
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
		log.Println("current options:")
		for k, v := range opts {
			log.Printf("  %s: '%s'\n", k, v.Val)
		}

		log.Println("loaded scripts:")
		for _, v := range sandbox.Scripts() {
			log.Printf("%s @ %s\n", v.Name(), v.Path())
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

func broadcast() Command {
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

func debug() Command {
	cmd := newCommand("enables debug logging", "debug")
	cmd.fn = func(line string, parts ...string) {
		if store.Debug {
			log.Println("disabled debugging")
		} else {
			log.Println("enabled debugging")
		}
		store.Debug = !store.Debug
	}
	return cmd
}

func reload() Command {
	cmd := newCommand("reloads the base.js", "reload")
	cmd.fn = func(line string, parts ...string) {
		sandbox.ReloadBase()
	}
	return cmd
}

func getUser() noye.User {
	return irc.ParseUser(fmt.Sprintf("%s!user@localhost", opts["from"]))
}
