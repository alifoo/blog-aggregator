package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/alifoo/blog-aggregator/internal/config"
	"github.com/alifoo/blog-aggregator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

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

func handlerReset(s *state, cmd command) error {
	if cmd.arguments == nil {
		return errors.New("no arguments passed to the handler")
	}
	err := s.db.DeleteAllUsers(context.Background())
	if err != nil {
		fmt.Printf("Error deleting all users:\n %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Deleted all users successfully.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	if cmd.arguments == nil {
		return errors.New("no arguments passed to the handler")
	}
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Error getting all users:\n %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}


	fmt.Println("All users:")
	for _, u := range users {
		if u.Name == cfg.CurrentUserName {
			u.Name = u.Name + " (current)"
		}
		fmt.Println(u.Name)
	}

	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request for url: %v\n %v", feedURL, err)
	}

	req.Header.Set("User-Agent", "gator")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error fetching url: %v\n %v", feedURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %v\n %v", feedURL, err)
	}

	var rssFeed RSSFeed
	if err := xml.Unmarshal(body, &rssFeed); err != nil {
		return nil, fmt.Errorf("error unmarshalling body: %v\n %v", feedURL, err)
	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for i := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}

	return &rssFeed, nil
}

func handlerAgg(s *state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error fetching feed: %v", err)
	}
	fmt.Println(feed)
	return nil
}

func handlerAddFeed(s *state, cmd command) error {
	if len(cmd.arguments) < 2 {
		return errors.New("not enough arguments passed to the handler")
	}
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	user, err := s.db.GetUser(context.Background(), cfg.CurrentUserName)
    if err != nil {
        if err == sql.ErrNoRows {
            fmt.Printf("User '%s' does not exist\n", cfg.CurrentUserName)
            os.Exit(1)
        }
        return fmt.Errorf("database error: %w", err)
	}

	name := cmd.arguments[0]
	url := cmd.arguments[1]

	params := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name: name,
		Url: url,
		UserID: user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params); if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}

	fmt.Println(feed)
	return nil
}

//func handlerFeeds(s *state, cmd command) error {
//
//}

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
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddFeed)

	err = commands.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
