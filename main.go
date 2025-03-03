package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/alifoo/blog-aggregator/internal/config"
)

type state struct {
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
		return fmt.Errorf("Command %s not found", cmd.name)
	}
	return handler(s, cmd)
}


func handlerLogin(s *state, cmd command) error {
	if cmd.arguments == nil {
		return errors.New("No arguments passed to the handler")
	}

	err := s.configPointer.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("The user %v has been set.", cmd.arguments[0])

	return nil
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var s state
	s.configPointer = &cfg


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

	commands.register("login", func(s *state, cmd command) error {
		if len(cmd.arguments) == 0 {
			return fmt.Errorf("a username is required for the login command")
		}
		username := cmd.arguments[0]
		s.configPointer.SetUser(username)
		fmt.Printf("Logged in as %s\n", username)
    	return nil
	})

	err = commands.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
