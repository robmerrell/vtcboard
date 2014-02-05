package updaters

import (
	"github.com/SlyMarbo/rss"
	"github.com/robmerrell/vtcboard/models"
)

type Reddit struct{}

// Update retrieves any new stores on the vertcoin subreddit.
func (r *Reddit) Update() error {
	conn := models.CloneConnection()
	defer conn.Close()

	vertcoinPosts, err := getNewRedditPosts("http://www.reddit.com/r/vertcoin/.rss", "/r/vertcoin")
	if err != nil {
		return err
	}

	if err := savePosts(vertcoinPosts, conn); err != nil {
		return err
	}

	vertmarketPosts, err := getNewRedditPosts("http://www.reddit.com/r/vertmarket/.rss", "/r/vertmarket")
	if err != nil {
		return err
	}

	return savePosts(vertmarketPosts, conn)
}

func savePosts(posts []*models.Post, conn *models.MgoConnection) error {
	for _, post := range posts {
		if err := post.Insert(conn); err != nil {
			return err
		}
	}

	return nil
}

func getNewRedditPosts(feedUrl, source string) ([]*models.Post, error) {
	conn := models.CloneConnection()
	defer conn.Close()

	// get the feed
	feed, err := rss.Fetch(feedUrl)
	if err != nil {
		return []*models.Post{}, err
	}

	// iterate the feed and save the new items
	posts := make([]*models.Post, 0)
	for _, item := range feed.Items {
		exists, err := models.PostExists(conn, item.ID)
		if err != nil {
			return []*models.Post{}, err
		}

		if !exists {
			post := &models.Post{
				Title:       item.Title,
				Source:      source,
				Url:         item.Link,
				UniqueId:    item.ID,
				PublishedAt: item.Date,
			}

			posts = append(posts, post)
		}
	}

	return posts, nil
}
