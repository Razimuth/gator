// handler_user.go
package main

import (
	"context"
	"fmt"
	"log"
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
	url := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("Error: %w", err)
	}
	// Print the entire feed struct
	fmt.Printf("%+v\n", feed)

	// Print in a readable format
	//	fmt.Printf("Channel: %s\nDescription: %s\nItems:\n", feed.Channel.Title, feed.Channel.Description)
	//	for _, item := range feed.Channel.Item {
	//		fmt.Printf("- %s\n", item.Title)
	//		fmt.Printf("Description: %s\n", item.Description)
	//	}
	return nil
}
