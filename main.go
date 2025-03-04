package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/alifoo/blog-aggregator/internal/config"
	"github.com/alifoo/blog-aggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type state struct {
	db *database.Queries
	configPointer *config.Config
}

type command struct {
	name string
	arguments []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("command %s not found", cmd.name)
	}
	return handler(s, cmd)
}


func handlerLogin(s *state, cmd command) error {
	if cmd.arguments == nil {
		return errors.New("no arguments passed to the handler")
	}

	username := cmd.arguments[0]
	_, err := s.db.GetUser(context.Background(), username)
    if err != nil {
        if err == sql.ErrNoRows {
            fmt.Printf("User '%s' does not exist\n", username)
            os.Exit(1)
        }
        return fmt.Errorf("database error: %w", err)
    }

	err = s.configPointer.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Printf("The user %v has been set.\n", username)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if cmd.arguments == nil {
		return errors.New("no arguments passed to the handler")
	}

	name := cmd.arguments[0]

	params := database.CreateUserParams{
        ID:        uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name:      name,
    }

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			fmt.Println("user already exists: %w", err)
			os.Exit(1)
		}
		fmt.Println(err)
		os.Exit(1)
	}

	err = s.configPointer.SetUser(name)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("The user was successfully created. User data:")
	fmt.Println(user)

	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	var s state
	s.configPointer = &cfg
	s.db = dbQueries

	commands := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	if len(os.Args) < 2 {
		fmt.Println("No arguments passed to the command")
		os.Exit(1)
	}

	cmd := command{
		name: os.Args[1],
		arguments: os.Args[2:],
	}

	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)

	err = commands.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
