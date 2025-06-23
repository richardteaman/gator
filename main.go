package main

import (
	"errors"
	"fmt"
	"gator/internal/config"
	"log"
	"os"
)

type state struct {
	Config *config.Config
}

type command struct {
	Name string
	Args []string
}

type commands struct {
	list map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.list[cmd.Name]
	if !exists {
		return fmt.Errorf("command not found :%v", cmd.Name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	if c.list == nil {
		c.list = make(map[string]func(*state, command) error)
	}
	c.list[name] = f
}

func main() {
	cfg, err := config.ReadConfigFile()
	if err != nil {
		log.Fatal("Error reading config: ", err)
	}

	appState := state{
		Config: &cfg,
	}

	appCommands := commands{
		list: make(map[string]func(*state, command) error),
	}
	appCommands.register("login", handlerLogin)

	if len(os.Args) < 2 {
		log.Fatal("no command provided")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	cmd := command{
		Name: cmdName,
		Args: cmdArgs,
	}

	err = appCommands.run(&appState, cmd)
	if err != nil {
		log.Fatal("command failed: ", err)
	}

	//fmt.Printf("Current config: %+v\n", appState.Config)
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return errors.New("no arguments provided")
	}
	username := cmd.Args[0]
	err := s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("login username set error: %v", err)
	}
	fmt.Println("user has been set")

	return nil

}
