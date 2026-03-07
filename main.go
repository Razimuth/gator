// main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Razimuth/gator/internal/config"
	"github.com/Razimuth/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

func main() {
	// Check if a command was provided
	if len(os.Args) < 2 {
		fmt.Println("Error: not enough arguments were provided.")
		os.Exit(1)
	}
	// Read the configuration file
	//	fmt.Println("Attempting to read config file...")

	cfgVal, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config %v\n", err)
		os.Exit(1)
	}
	//	fmt.Printf("Config data: %+v\n", cfgVal)

	db, err := sql.Open("postgres", cfgVal.DBURL)
	if err != nil {
		log.Fatal("unable to open database connection:", err)
	}
	dbQueries := database.New(db)
	defer db.Close()

	// create the application state
	appState := &state{db: dbQueries, cfg: &cfgVal} // pointer to state, holding pointer to cfg

	// create the commands manager and register handlers
	cmdManager := commands{
		handlers: make(map[string]func(*state, command) error),
	}
	cmdManager.register("login", handlerLogin)
	cmdManager.register("register", handlerRegister)
	cmdManager.register("reset", handlerReset)
	cmdManager.register("users", handlerUsers)
	cmdManager.register("agg", handlerAgg)
	cmdManager.register("addfeed", handlerAddFeed)
	cmdManager.register("feeds", handlerFeeds)
	cmdManager.register("follow", handlerFollow)
	cmdManager.register("following", handlerFollowing)

	// Register other commands here later

	// command-line arguments passes into a command struct
	cmdName := os.Args[1]
	cmdArgs := os.Args[2:] // Slice from the third argument onwards
	cmdInstance := command{
		Name: cmdName,
		Args: cmdArgs,
	}

	// Run the command
	if err := cmdManager.run(appState, cmdInstance); err != nil {
		fmt.Printf("Command error: %v\n", err)
		os.Exit(1)
	}
}
