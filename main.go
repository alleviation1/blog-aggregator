package main

import (
	"os"
	"database/sql"
	"log"
	"context"
	"fmt"
	
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
	c.register("register", handlerRegister)
	c.register("login", handlerLogin)
	c.register("reset", handlerReset)
	c.register("users", handlerGetUsers)
	c.register("agg", handlerAggregate)
	c.register("feeds", handlerGetFeeds)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("browse", middlewareLoggedIn(handlerBrowse))

	if len(os.Args) < 2 {
		log.Fatalf("Expected at least 2 arguments")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]


	if err := c.run(s, command{name: cmdName, args: cmdArgs}); err != nil {
		log.Fatalf("Error: %w", err)
	}

}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.config.User)
		if err != nil {
			log.Fatalf("Unable to log user in within middlewareLoggedIn")
		}

		err = handler(s, cmd, user)
		if err != nil {
			return fmt.Errorf("Unable to complete handler within middleware login: %w", err)
		}
		return nil
	}
}