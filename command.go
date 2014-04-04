package main

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
