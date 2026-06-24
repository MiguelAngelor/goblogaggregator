package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"github.com/MiguelAngelor/goblogaggregator/internal/config"
	"github.com/MiguelAngelor/goblogaggregator/internal/database"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"html"
	"io"
	"net/http"
	"os"
	"time"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

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

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command")
	}

	return handler(s, cmd)

}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username required")
	}

	username := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), username)
	if err != nil {
		return fmt.Errorf("user does not exist")
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Printf("User set to %s\n", username)

	return nil

}

func main() {

	//*

	cfg, err := config.Read()
	if err != nil {
		panic(err)
	}

	DbUrl := cfg.DbUrl

	db, err := sql.Open("postgres", DbUrl)
	if err != nil {
		panic(err)
	}

	dbQueries := database.New(db)

	s := state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
cmds.register("follow", middlewareLoggedIn(handlerFollow))
cmds.register("following", middlewareLoggedIn(handlerFollowing))
cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))


	args := os.Args

	if len(args) < 2 {
		fmt.Println("no command provided")
		os.Exit(1)
	}

	cmd := command{
		name: args[1],
		args: args[2:],
	}

	err = cmds.run(&s, cmd)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func middlewareLoggedIn(
	handler func(s *state, cmd command, user database.User) error,
) func(*state, command) error {

	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(
			context.Background(),
			s.cfg.CurrentUserName,
		)
		if err != nil {
			return err
		}

		return handler(s, cmd, user)
	}
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username required")
	}

	name := cmd.args[0]

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})

	if err != nil {
		return err
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}

	fmt.Println("User created successfully!")
	fmt.Println(user)

	return nil

}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Println("Database reset succesful")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil

}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var feed RSSFeed

	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, err
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title =
			html.UnescapeString(feed.Channel.Item[i].Title)

		feed.Channel.Item[i].Description =
			html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil

}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(
		context.Background(),
	)
	if err != nil {
		return err
	}

	fmt.Printf("Fetching feed: %s\n", feed.Name)

	err = s.db.MarkFeedFetched(
		context.Background(),
		feed.ID,
	)
	if err != nil {
		return err
	}

	rssFeed, err := fetchFeed(
		context.Background(),
		feed.Url,
	)
	if err != nil {
		return err
	}

	for _, item := range rssFeed.Channel.Item {
    // create post here
	}

	return nil
}


func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("duration required")
	}

	timeBetweenRequests, err :=
		time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf(
		"Collecting feeds every %v\n",
		timeBetweenRequests,
	)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		err := scrapeFeeds(s)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	feed, err := s.db.CreateFeed(
		context.Background(),
		database.CreateFeedParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Name:      name,
			Url:       url,
			UserID:    user.ID,
		},
	)
	if err != nil {
		return err
	}

	_, err = s.db.CreateFeedFollow(
	context.Background(),
	database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
		},
	)

	if err != nil {
		return err
	}

	fmt.Println(feed)

	return nil

}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf(
			"Name: %s\nUrl: %s\nUser: %s\n\n",
			feed.Name,
			feed.Url,
			feed.UserName,
		)
	}

	return nil

}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("url required")
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(
		context.Background(),
		url,
	)
	if err != nil {
		return err
	}

	follow, err := s.db.CreateFeedFollow(
		context.Background(),
		database.CreateFeedFollowParams{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    user.ID,
			FeedID:    feed.ID,
		},
	)
	if err != nil {
		return err
	}

	fmt.Printf("%s is now following %s\n",
		follow.UserName,
		follow.FeedName,
	)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
    follows, err := s.db.GetFeedFollowsForUser(
        context.Background(),
        	user.Name,
    )
    if err != nil {
        return err
    }

    for _, follow := range follows {
        fmt.Println(follow.FeedName)
    }

    return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("url required")
	}

	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(
		context.Background(),
		url,
	)
	if err != nil {
		return err
	}

	err = s.db.DeleteFeedFollow(
		context.Background(),
		database.DeleteFeedFollowParams{
			UserID: user.ID,
			FeedID: feed.ID,
		},
	)
	if err != nil {
		return err
	}

	fmt.Println("Unfollowed", feed.Name)

	return nil
}
//* Commands
