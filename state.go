package main

import (
	"fmt"
	"errors"
	"time"
	"context"
	"html"
	"database/sql"
	"strconv"

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

func handlerGetUsers(s *state, cmd command) error {
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

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("aggregate expects 1 argument: <time between requests>\n")
	}

	time_between_reqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Unable to create ticker in aggreagte: %w", err)
	}

	fmt.Printf("Collecting reqests every %v\n", time_between_reqs)

	ticker := time.NewTicker(time_between_reqs)

	for ; ; <- ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			return fmt.Errorf("Error in aggregate handler: %w", err)
		}
	}

	return nil
	// feed, err := fetchFeeds(context.Background(), "https://www.wagslane.dev/index.xml")
	// if err != nil {
	// 	return err
	// }

	// 	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	// 	feed.Channel.Link = html.UnescapeString(feed.Channel.Link)
	// 	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	// 	for i, item := range feed.Channel.Item {
	// 		item.Title = html.UnescapeString(item.Title)
	// 		item.Link = html.UnescapeString(item.Link)
	// 		item.Description = html.UnescapeString(item.Description)
	// 		item.PubDate = html.UnescapeString(item.PubDate)
	// 		feed.Channel.Item[i] = item
	// 	}

	// 	fmt.Printf("Feed: %+v\n", *feed)

	// return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Add feed expected 2 arguments in command: %w\n")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams {
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})

	if err != nil {
		return fmt.Errorf("Unable to create feed: %w\n", err)
	}
	
	err = handlerFollow(s, command{name: "follow", args: []string{feed.Url}}, user)
	if err != nil {
		return fmt.Errorf("Unable to create feed follow in add feed: %w\n", err)
	}

	fmt.Printf("Feed:\n %+v\n", feed)
	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsUsers(context.Background())
	if err != nil {
		return fmt.Errorf("Unable to retrieve feeds in get feeds handler: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("Feed: %s\nUrl: %s\nUser: %v\n", feed.Name, feed.Url, feed.Name_2)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Follow handler only expects 1 command argument <url>\n")
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Unable to get feed in handler follow: %w\n", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams {
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Unable to create feed follow: %w", err)
	}

	fmt.Printf("Feed: %v\nUser: %v\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedsFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Unable to get feed follows for user in following handler: %w", err)
	}

	fmt.Printf("Feeds for %v:\n", user.Name)

	for _, feedsFollow := range feedsFollows {
		fmt.Printf("FeedFollow: %+v\n", feedsFollow)
		feed, err := s.db.GetFeedByID(context.Background(), feedsFollow.FeedID)
		if err != nil {
			return fmt.Errorf("Error getting feed in feed follows: %w\n", err)
		}
		fmt.Printf("FeedFollow Feed Name: %v\n", feed.Name)
	}
	return nil
}

func handlerUnfollow(s* state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Unfollow expects 1 argument <feed url>\n")
	}

	url := cmd.args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Unable to get feed by url in unfollow: %w", err)
	}

	err = s.db.DeleteFeedFollowsForUser(context.Background(), database.DeleteFeedFollowsForUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})

	if err != nil {
		return fmt.Errorf("Unable to delete feed for user in unfollow: %w", err)
	}

	fmt.Println("Unfollow successful")
	return nil
}

func scrapeFeeds(s *state) error {
	// get next feed
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Unable to get next feed in scrapeFeed: %w", err)
	}

	fmt.Printf("Feed: %v\n", feed.Name)

	// mark fetched
	feed, err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return fmt.Errorf("Unable to mark feed in scrapeFeed: %w", err)
	}

	fmt.Printf("Feed: %v successfully marked\n", feed.Name)

	// fetch feed(url)
	rssFeed, err := fetchFeeds(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("Unable to fetch feeds in scrapeFeed: %w", err)
	}

	// iterate and print item titles in feed
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)

	// YYYY-MM-DD hh:mm:ss
	for _, item := range rssFeed.Channel.Item {
		parsedTime, err := parsePublishingDate(item.PubDate)
		if err != nil {
			return fmt.Errorf("Error parsing publishing date in scrape feeds: %w\n", err)
		}
		post, err := s.db.CreatePost(context.Background(), database.CreatePostParams {
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       html.UnescapeString(item.Title),
			Url:         html.UnescapeString(item.Link),
			Description: sql.NullString {String: html.UnescapeString(item.Description), Valid: true,},
			PublishedAt: parsedTime,
			FeedID:      feed.ID,
		})
		if err != nil {
			return fmt.Errorf("Error creating post from channel item in scrape feeds: %w\n", err)
		}
		fmt.Printf("Post Created: %v\n", post.Title)
	}

	fmt.Printf("Feed: %+v\n", *rssFeed)

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) == 1 {
		newLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("Unable to parse int in handler browse: %w\n", err)
		} else {
			limit = newLimit
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("Couldn't get posts for user in handler browse: %w\n", err)
	}

	fmt.Println("Posts:")
	for _, post := range posts {
		fmt.Printf("Post: %+v\n", post)
	}
	return nil
}

func parsePublishingDate(date string) (sql.NullTime, error) {
	escapedDate := html.UnescapeString(date)
	parsedTime, err := time.Parse(time.RFC1123Z, escapedDate)
	if err != nil {
		return sql.NullTime{Time: time.Time{}, Valid: false}, err
	}

	return sql.NullTime {
		Time: parsedTime,
		Valid: true,
	}, nil
}