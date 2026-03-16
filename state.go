package main

import (
	"fmt"
	"errors"

	"github.com/alleviation1/blog_aggregator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commands map[string]func(*state, command) error
}


func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	f, ok := c.commands[cmd.name]
	if !ok {
		return errors.New("command not found")
	}
	return f(s, cmd)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Login handler expected a single argument <username> to be passed")
	}

	user := cmd.args[0]
	if err := s.config.SetUser(user); err != nil {
		return fmt.Errorf("Error setting user: %w", err)
	}

	fmt.Printf("User set: %s\n", s.config.User)
	return nil
}