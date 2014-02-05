package updaters

import (
	"github.com/SlyMarbo/rss"
	"github.com/robmerrell/vtcboard/models"
)

var subredditUrl = "http://www.reddit.com/r/worldcoin/.rss"

type Reddit struct{}

// Update retrieves any new stores on the worldcoin subreddit.
func (r *Reddit) Update() error {
	conn := models.CloneConnection()
	defer conn.Close()

	posts, err := getNewRedditPosts()
	if err != nil {
		return err
	}

	for _, post := range posts {
		if err := post.Insert(conn); err != nil {
			return err
		}
	}

	return nil
}

func getNewRedditPosts() ([]*models.Post, error) {
	conn := models.CloneConnection()
	defer conn.Close()

	// get the feed
	feed, err := rss.Fetch(subredditUrl)
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
				Source:      "reddit",
				Url:         item.Link,
				UniqueId:    item.ID,
				PublishedAt: item.Date,
			}

			posts = append(posts, post)
		}
	}

	return posts, nil
}
