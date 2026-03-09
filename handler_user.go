// handler_user.go
package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Razimuth/gator/internal/config"
	"github.com/Razimuth/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("login command requires username")
	}
	username := cmd.Args[0]

	_, err := s.db.GetUser(context.Background(), username)

	// If no rrors returned, the user already exists
	if err != nil {
		return fmt.Errorf("Error finding username: %s: %v\n", username, err)
	}

	// Update the config within the state using SetUser
	s.cfg.SetUser(username)

	fmt.Printf("Username set to: %s\n", username)

	fmt.Println("Attempting to read config file...")

	cfgVal, err := config.Read()
	if err != nil {
		return fmt.Errorf("Error reading config %v\n", err)
	}
	fmt.Printf("Config data: %+v\n", cfgVal)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("a name is required to register")
	}
	name := cmd.Args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: name})

	if err != nil {
		return fmt.Errorf("Error creating user: %w", err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("could not set current user: %w", err)
	}

	fmt.Printf("User created: %s (%v)\n", user.Name, user.ID)
	log.Printf("User data: %+v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not reset users: %w", err)
	}
	fmt.Println("All users have been reset.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not get users: %w", err)
	}
	for _, user := range users {
		currentUser := s.cfg.CurrentUserName
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("need <time_between_reqs>, 1m (Runs every minute).suggest > 1m")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	defer ticker.Stop()

	// Run immediately, then on tick
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.Args) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.Args[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("- %s\n", post.Title)
		fmt.Printf(" Link: %s\n", post.Url)
		fmt.Printf(" from %s\n", post.FeedName)
		fmt.Printf(" (Published: %s)\n", post.PublishedAt.Time.Format(time.RFC1123))
		fmt.Printf(" Description: %+v\n", post.Description)
		fmt.Println("=====================================")
	}
	return nil
}
