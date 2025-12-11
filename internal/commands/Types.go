package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jaxhemopo/rssgator/internal/config"
	"github.com/jaxhemopo/rssgator/internal/database"
	"github.com/jaxhemopo/rssgator/internal/rss"
)

type State struct {
	DB     *database.Queries
	Config *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Handlers map[string]func(*State, Command) error
}

const GatorURLfeed = "https://www.wagslane.dev/index.xml"

func MiddlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		if s.Config.CurrentUserName == "" {
			fmt.Printf("no user logged in. Please login first.\n")
			os.Exit(1)
		}
		user, err := s.DB.GetUser(context.Background(), s.Config.CurrentUserName)
		if err != nil {
			fmt.Printf("failed to get user: %v", err)
			os.Exit(1)
		}
		return handler(s, cmd, user)
	}
}

func (c *Commands) Run(s *State, cmd Command) error {
	if handler, ok := c.Handlers[cmd.Name]; !ok {
		return fmt.Errorf("unknown Command: %s", cmd.Name)
	} else {
		err := handler(s, cmd)
		if err != nil {
			return fmt.Errorf("error executing command %s: %v", cmd.Name, err)
		}
	}
	return nil
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Handlers[name] = f
}

func AddFeedHandler(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 2 {
		fmt.Printf("feed name and url required")
		os.Exit(1)
	}
	id := uuid.New()
	timeNow := time.Now()
	feedname := cmd.Args[0]
	feedurl := cmd.Args[1]

	params := database.CreateFeedParams{
		ID:        id,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Name:      feedname,
		Url:       feedurl,
		UserID:    user.ID,
	}

	_, err := s.DB.CreateFeed(context.Background(), params)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			fmt.Printf("Feed %s already exists\n", cmd.Args[0])
			os.Exit(1)
		}
		fmt.Printf("failed to create feed: %v", err)
		return err
	}

	createFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		FeedID:    params.ID,
		UserID:    user.ID,
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), createFollowParams)
	if err != nil {
		fmt.Printf("failed to create feed follow: %v", err)
		return err
	} else {
		fmt.Printf("You are now following feed %s\n", feedname)
	}

	fmt.Printf("Feed %s added successfully\n", cmd.Args[0])
	fmt.Printf("%d", params.ID)
	fmt.Printf("%s", params.Name)
	fmt.Printf("%s", params.Url)
	fmt.Printf("%s", params.CreatedAt)
	fmt.Printf("%s", params.UpdatedAt)
	fmt.Printf("%d", params.UserID)

	return nil
}

func FeedsHandler(s *State, cmd Command) error {
	feeds, err := s.DB.ListAllFeeds(context.Background())
	if err != nil {
		fmt.Printf("failed to get feeds: %v", err)
		os.Exit(1)
	}
	if len(feeds) == 0 {
		fmt.Printf("No feeds found\n")
		return nil
	}
	fmt.Printf("Registered Feeds:\n")
	for _, feed := range feeds {
		user, err := s.DB.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			fmt.Printf("failed to get user for feed %s: %v", feed.Name, err)
			continue
		}
		fmt.Printf("- '%s' : %s \n User: %s\n", feed.Name, feed.Url, user.Name)
	}
	return nil
}

func AggsHandler(s *State, cmd Command) error {
	feed, err := rss.FetchFeed(context.Background(), GatorURLfeed)
	if err != nil {
		fmt.Printf("failed to fetch feed: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s\n", feed.Channel.Title)
	fmt.Printf("%s\n", feed.Channel.Description)
	fmt.Printf("Items:\n")
	for _, item := range feed.Channel.Item {
		fmt.Printf("%s (%s) %s\n", item.Title, item.PubDate, item.Description)
	}
	return nil
}

func HandlerReset(s *State, cmd Command) error {
	err := s.DB.ResetUsers(context.Background())
	if err != nil {
		fmt.Printf("failed to reset users: %v", err)
		os.Exit(1)
	}
	fmt.Printf("All users have been reset successfully\n")
	return nil
}

func HandlerFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		fmt.Printf("URL required")
		os.Exit(1)
	}
	feedurl := cmd.Args[0]
	feed, err := s.DB.GetFeedByURL(context.Background(), feedurl)
	if err != nil {
		fmt.Printf("failed to get feed: %v", err)
		os.Exit(1)
	}

	timeNow := time.Now()
	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		FeedID:    feed.ID,
		UserID:    user.ID,
	}
	_, err = s.DB.CreateFeedFollow(context.Background(), params)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			fmt.Printf("You are already following feed %s\n", feed.Name)
			os.Exit(1)
		}
		fmt.Printf("failed to follow feed: %v", err)
		return err
	}
	fmt.Printf("%s is now following feed %s\n", user.Name, feed.Name)

	return nil
}

func HandlerFollowing(s *State, cmd Command, user database.User) error {
	feedFollows, err := s.DB.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		fmt.Printf("failed to get feed follows: %v", err)
		os.Exit(1)
	}
	if len(feedFollows) == 0 {
		fmt.Printf("You are not following any feeds\n")
		return nil
	}
	fmt.Printf("Feeds followed by %s:\n", user.Name)
	for _, feedFollow := range feedFollows {
		fmt.Printf("- '%s' \n", feedFollow.FeedName)
	}
	return nil
}

func HandlerUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Args) < 1 {
		fmt.Printf("Feed ID required")
		os.Exit(1)
	}
	feed, err := s.DB.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		fmt.Printf("failed to get feed: %v", err)
		os.Exit(1)
	}

	unfollowparams := database.DeleteFeedFollowParams{
		FeedID: feed.ID,
		UserID: user.ID,
	}

	err = s.DB.DeleteFeedFollow(context.Background(), unfollowparams)
	if err != nil {
		fmt.Printf("failed to unfollow feed: %v", err)
		os.Exit(1)
	}
	fmt.Printf("%s has unfollowed feed %s\n", user.Name, feed.Name)
	return nil
}

func HandlerGetUsers(s *State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("failed to get users: %v", err)
		os.Exit(1)
	}
	if len(users) == 0 {
		fmt.Printf("No users found\n")
		return nil
	}
	fmt.Printf("Registered Users:\n")
	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("- '%s (current)'\n", user.Name)
			continue
		} else {
			fmt.Printf("- '%s' \n", user.Name)
		}
	}
	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		fmt.Printf("username required")
		os.Exit(1)
	}
	_, err := s.DB.GetUser(context.Background(), cmd.Args[0])
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			fmt.Printf("User %s does not exist\n", cmd.Args[0])
			os.Exit(1)
		}
		fmt.Printf("failed to get user: %v", err)
		os.Exit(1)
	}

	err = s.Config.SetUser(cmd.Args[0])
	if err != nil {
		fmt.Printf("failed to set user: %v", err)
		os.Exit(1)
	}
	fmt.Printf("User set to %s\n", cmd.Args[0])
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		return fmt.Errorf("username required")
	}
	id := uuid.New()
	timeNow := time.Now()

	params := database.CreateUserParams{
		ID:        id,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
		Name:      cmd.Args[0],
	}

	createdUser, err := s.DB.CreateUser(context.Background(), params)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			fmt.Printf("User %s already exists\n", cmd.Args[0])
			os.Exit(1)
		}
		fmt.Printf("failed to create user: %v", err)
		return err
	}
	s.Config.SetUser(createdUser.Name)
	fmt.Printf("User %s registered successfully\n", createdUser.Name)
	return nil
}
