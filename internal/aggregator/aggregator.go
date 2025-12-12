package aggregator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jaxhemopo/rssgator/internal/database"
	"github.com/jaxhemopo/rssgator/internal/rss"
)

func ScrapeFeeds(ctx context.Context, db *database.Queries) error {
	feed, err := db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	fetchedTime := sql.NullTime{
		Time:  now.UTC(),
		Valid: true,
	}
	err = db.MarkFeedFetched(ctx, database.MarkFeedFetchedParams{
		LastFetchedAt: fetchedTime,
		UpdatedAt:     now,
		ID:            feed.ID,
	})
	if err != nil {
		return err
	}
	rssFeed, err := rss.FetchFeed(ctx, feed.Url)
	if err != nil {
		return err
	}

	fmt.Printf("%s\n", rssFeed.Channel.Title)
	fmt.Printf("Items:\n")
	for _, item := range rssFeed.Channel.Item {
		pubDate, err := rss.ParsePubDate(item.PubDate)
		if err != nil {
			return err
		}
		fmt.Printf("- %s (%s)\n", item.Title, item.Link)
		_, err = db.CreatePost(ctx, database.CreatePostParams{
			ID:        uuid.New(),
			CreatedAt: now,
			UpdatedAt: now,
			FeedID:    feed.ID,
			Title:     item.Title,
			Url:       item.Link,
			Description: sql.NullString{String: item.Description,
				Valid: true,
			},
			PublishedAt: pubDate,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Printf("  Post already exists: %s\n", item.Title)
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		} else {
			log.Printf("  Added post: %s\n", item.Title)
		}
	}
	return nil
}
