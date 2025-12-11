package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/jaxhemopo/rssgator/internal/commands"
	"github.com/jaxhemopo/rssgator/internal/config"
	"github.com/jaxhemopo/rssgator/internal/database"
	_ "github.com/lib/pq"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: missing command name. Usage: gator <command> [args...]\n")
		os.Exit(1)
	}

	cliCmd := commands.Command{
		Name: os.Args[1],
		Args: os.Args[2:],
	}

	cfg, err := config.Read()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	dbQueries := database.New(db)

	appState := &commands.State{
		Config: &cfg,
		DB:     dbQueries,
	}

	cmdRegistry := &commands.Commands{
		Handlers: make(map[string]func(*commands.State, commands.Command) error),
	}

	cmdRegistry.Register("login", commands.HandlerLogin)
	cmdRegistry.Register("register", commands.HandlerRegister)
	cmdRegistry.Register("reset", commands.HandlerReset)
	cmdRegistry.Register("users", commands.HandlerGetUsers)
	cmdRegistry.Register("agg", commands.AggsHandler)
	cmdRegistry.Register("addfeed", commands.MiddlewareLoggedIn(commands.AddFeedHandler))
	cmdRegistry.Register("feeds", commands.FeedsHandler)
	cmdRegistry.Register("follow", commands.MiddlewareLoggedIn(commands.HandlerFollow))
	cmdRegistry.Register("following", commands.MiddlewareLoggedIn(commands.HandlerFollowing))
	cmdRegistry.Register("unfollow", commands.MiddlewareLoggedIn(commands.HandlerUnfollow))

	err = cmdRegistry.Run(appState, cliCmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

}
