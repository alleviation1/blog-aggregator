package main

import (
	"fmt"
	"os"

	"github.com/alleviation1/blog_aggregator/internal/config"
)

const configFileName = "/gatorconfig.json"

func main() {
	cfg, err := config.Read(configFileName)
	if err != nil {
		fmt.Errorf("Error creating config: %w", err)
		return
	}
	fmt.Printf("Config before operations: %+v\n", cfg)


	if len(os.Args) < 2 {
		fmt.Println("Program requires at least 2 arguments")
		os.Exit(1)
	}

	args := os.Args[1:]
	if len(args) < 2 {
		fmt.Println("Login requires a username to be passed")
		os.Exit(1)
	}

	s := state{config: &cfg}

	c := commands{commands: make(map[string]func(*state, command) error)}

	login := command{
		name: args[0],
		args: args[1:],
	}

	c.register(login.name, handlerLogin)

	if err := c.run(&s, login); err != nil {
		fmt.Printf("Error running function: %w", err)
		return
	}

}