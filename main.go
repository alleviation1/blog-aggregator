package main

import (
	"os"
	"database/sql"
	"log"
	
	_ "github.com/lib/pq"
	"github.com/alleviation1/blog_aggregator/internal/config"
	"github.com/alleviation1/blog_aggregator/internal/database"
)

const configFileName = "/gatorconfig.json"

func main() {
	cfg, err := config.Read(configFileName)
	if err != nil {
		log.Fatalf("Error creating config: %w", err)
	}

	s := &state{config: &cfg}

	db, err := sql.Open("postgres", s.config.Url)
	defer db.Close()
	s.db = database.New(db)

	c := commands{commands: make(map[string]func(*state, command) error)}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerGetUsers)
	c.register("agg", handlerAggregate)

	if len(os.Args) < 2 {
		log.Fatalf("Expected at least 2 arguments")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]


	if err := c.run(s, command{name: cmdName, args: cmdArgs}); err != nil {
		log.Fatalf("Error: %w", err)
	}

}