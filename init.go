package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/j6n/noye/ext"
	"github.com/j6n/noye/mock"
	"github.com/j6n/noye/noye"
)

var (
	commands map[string]Command
	sandbox  noye.Manager
	scanner  = bufio.NewScanner(os.Stdin)
	output   = make(chan string)

	opts = map[string]*option{
		"from": &option{"from", "test", "nick of person sending the messages"},
		"chan": &option{"chan", "#noye", "the channel which messages are sent to"},
	}
)

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
