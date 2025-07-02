package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"gator/internal/config"
	"gator/internal/database"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type state struct {
	Config *config.Config
	db     *database.Queries
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

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatal("can't connect to db", err)
	}
	defer db.Close()

	dbQueries := database.New(db)

	appState := state{
		Config: &cfg,
		db:     dbQueries,
	}

	appCommands := commands{
		list: make(map[string]func(*state, command) error),
	}
	appCommands.register("login", handlerLogin)
	appCommands.register("register", handlerRegister)
	appCommands.register("reset", handlerReset)
	appCommands.register("users", handlerUsers)

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

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("User not found")
			os.Exit(1)
		}

		fmt.Printf("Error getting user: %v\n", err)
		os.Exit(1)
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("login username set error: %v", err)
	}
	fmt.Println("user has been set")

	return nil

}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) < 1 {
		return errors.New("no arguments provided")
	}

	username := cmd.Args[0]
	userID := uuid.New()
	now := time.Now()

	user, err := s.db.CreateUser(
		context.Background(),
		database.CreateUserParams{
			ID:        userID,
			CreatedAt: now,
			UpdatedAt: now,
			Name:      username,
		},
	)
	if err != nil {
		if isUniqueViolation(err) {
			fmt.Println("User already exists")
			os.Exit(1)
		}
		return fmt.Errorf("could not create user: %v", err)
	}

	err = s.Config.SetUser(username)
	if err != nil {
		return fmt.Errorf("could not set user in config file: %v", err)
	}

	fmt.Printf("User registered: \nID=%v, \nName=%v\nCreatedAt=%v\nUpdatedAt=%v\n", user.ID, user.Name, user.CreatedAt, user.UpdatedAt)
	return nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "already exists"))
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not reset the app: %v", err)
	}
	fmt.Println("app reset")

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not get users from db", err)
	}
	if len(users) < 1 {
		return fmt.Errorf("no users found")
	}

	currentUser := s.Config.CurrentUserName
	for _, user := range users {
		if user.Name == currentUser {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}

	}
	return nil
}
