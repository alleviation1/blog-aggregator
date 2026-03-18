package main

import (
	"fmt"
	"errors"
	"time"
	"context"
	"html"

	"github.com/google/uuid"
	"github.com/alleviation1/blog_aggregator/internal/config"
	"github.com/alleviation1/blog_aggregator/internal/database"

)

type state struct {
	config *config.Config
	db	   *database.Queries
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
		return errors.New("Login handler expected a single argument <username> to be passed")
	}

	name := cmd.args[0]

	user, err := s.db.GetUser(context.Background(), name)
	if err != nil {
		return fmt.Errorf("Error getting user: %w\n", err)
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("Error setting user in login handler: %w\n", err)
	}

	fmt.Printf("User set: %s\n", s.config.User)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Command must contain a name")
	}

	name := cmd.args[0]

	if _, err := s.db.GetUser(context.Background(), name); err == nil {
		return errors.New("user already exists")
	}

	newUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: 		 uuid.New(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Name:	     name,
	})
	
	if err != nil {
		return fmt.Errorf("Unable to create user %w\n", err)
	}


	err = s.config.SetUser(newUser.Name)
	if err != nil {
		return fmt.Errorf("Couldn't set user in register handler: %w", err)
	}
	
	fmt.Println("user was created:")
	fmt.Printf("user data: %v+\n", newUser)

	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteUsers(context.Background()); err != nil {
		return fmt.Errorf("Couldn't delete users table data: %w", err)
	}

	fmt.Println("Users table data deleted")
	return nil
}

func handlerGetUsers(s* state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Couldn't retrieve users from users table: %w", err)
	}

	for _, user := range(users) {
		if user.Name == s.config.User {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAggregate(s* state, cmd command) error {
	feed, err := fetchFeeds(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

		feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
		feed.Channel.Link = html.UnescapeString(feed.Channel.Link)
		feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

		for i, item := range feed.Channel.Item {
			item.Title = html.UnescapeString(item.Title)
			item.Link = html.UnescapeString(item.Link)
			item.Description = html.UnescapeString(item.Description)
			item.PubDate = html.UnescapeString(item.PubDate)
			feed.Channel.Item[i] = item
		}

		fmt.Printf("Feed: %+v\n", *feed)

	return nil
}