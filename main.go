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
	appCommands.register("agg", handlerAgg)
	appCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	appCommands.register("feeds", handlerFeeds)
	appCommands.register("follow", middlewareLoggedIn(handlerFollow))
	appCommands.register("following", middlewareLoggedIn(handlerFollowing))
	appCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))

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

func handlerAgg(s *state, cmd command) error {
	url := "https://www.wagslane.dev/index.xml"

	rssFeed, err := fetchFeed(context.TODO(), url)
	if err != nil {
		log.Fatal("could not fetch rss from agg\n", err)
		return err
	}
	fmt.Printf("%+v\n", rssFeed)
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 2 {
		return errors.New("too few arguments provided")
	}
	name := cmd.Args[0]
	url := cmd.Args[1]

	user_id := user.ID

	now := time.Now()

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		feed_id := uuid.New()
		feed, err = s.db.CreateFeed(
			context.Background(),
			database.CreateFeedParams{
				ID:        feed_id,
				CreatedAt: now,
				UpdatedAt: now,
				Name:      name,
				Url:       url,
				UserID:    user_id,
			},
		)
		fmt.Printf("Feed created:\n")
		fmt.Printf("ID: %s\n", feed.ID)
		fmt.Printf("Name: %s\n", feed.Name)
		fmt.Printf("URL: %s\n", feed.Url)
		fmt.Printf("UserID: %s\n", feed.UserID)
		fmt.Printf("CreatedAt: %s\n", feed.CreatedAt)
		fmt.Printf("UpdatedAt: %s\n", feed.UpdatedAt)
		if err != nil {
			return fmt.Errorf("could not create feed: %v", err)
		}

	}

	_, err = s.db.GetFeedFollowForUserAndFeed(context.Background(),
		database.GetFeedFollowForUserAndFeedParams{
			UserID: user_id,
			FeedID: feed.ID,
		},
	)

	if errors.Is(err, sql.ErrNoRows) {
		feed_follow_id := uuid.New()
		_, err = s.db.CreateFeedFollow(
			context.Background(),
			database.CreateFeedFollowParams{
				ID:        feed_follow_id,
				CreatedAt: now,
				UpdatedAt: now,
				UserID:    user_id,
				FeedID:    feed.ID,
			},
		)
		if err != nil {
			return fmt.Errorf("could not create feed follow record: %v", err)
		}
		fmt.Printf("Followed feed with name: %v\n", feed.Name)
	} else if err != nil {
		return fmt.Errorf("could not check if feed follow exists: %v", err)
	} else {
		fmt.Println("feed is already followed by the user")
	}

	return nil
}

func handlerFeeds(s *state, cmd command) error {

	feeds, err := s.db.GetFeedsWithUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not get feeds from handlerFeeds: %v", err)
	}

	if len(feeds) == 0 {
		fmt.Println("No feeds found.")
		return nil
	}

	for _, feed := range feeds {
		fmt.Printf("Name: %v\n", feed.FeedName)
		fmt.Printf("URL: %v\n", feed.FeedUrl)
		fmt.Printf("User: %v\n", feed.UserName)
		fmt.Println("-----")
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("no arguments provided")
	}
	url := cmd.Args[0]
	feed_follow_id := uuid.New()
	now := time.Now()

	feed, err := s.db.GetFeedByURL(context.Background(), url)
	if err != nil {
		return fmt.Errorf("feed not found: %w", err)
	}

	feed_follows, err := s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        feed_follow_id,
			CreatedAt: now,
			UpdatedAt: now,
			UserID:    user.ID,
			FeedID:    feed.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create feed follow: %w", err)
	}

	fmt.Printf("Feed: %v\nUser: %v\n", feed_follows.FeedName, feed_follows.UserName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {

	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("feed following for this user not found: %w", err)
	}
	if len(following) < 1 {
		fmt.Println("Empty following")
	}

	for _, follow := range following {
		fmt.Printf("Feed name: %v\n", follow.FeedName)
	}

	return nil
}

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
) func(s *state, cmd command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("no user logged in. please log in first")
			}
			return fmt.Errorf("could not fetch current user: %w", err)
		}
		return handler(s, cmd, user)
	}
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.Args) < 1 {
		return errors.New("no arguments provided")
	}

	err := s.db.DeleteFeedFollow(
		context.Background(),
		database.DeleteFeedFollowParams{
			UserID: user.ID,
			Url:    cmd.Args[0],
		},
	)
	if err != nil {
		return fmt.Errorf("error deleting feed follow: %v", err)
	}

	return nil
}
