package updaters

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/robmerrell/wdcboard/models"
	"strings"
	"time"
)

var forumBaseUrl = "http://www.worldcoinforum.org/forum/%s/?sort_key=start_date&sort_by=Z-A"

type Forum struct{}

func (f *Forum) Update() error {
	conn := models.CloneConnection()
	defer conn.Close()

	discussionPosts, err := getNewTopicsFromForum("3-worldcoin-discussion")
	if err != nil {
		return err
	}
	for _, post := range discussionPosts {
		if err := post.Insert(conn); err != nil {
			return err
		}
	}

	promotionPosts, err := getNewTopicsFromForum("4-promotion-of-worldcoin")
	if err != nil {
		return err
	}
	for _, post := range promotionPosts {
		if err := post.Insert(conn); err != nil {
			return err
		}
	}

	return nil
}

// getNewTopicsFromForum takes in a forum url name and returns up to 5 of the latest topics not saved in
// the database.
func getNewTopicsFromForum(forumSuffix string) ([]*models.Post, error) {
	url := forumBaseUrl
	if strings.Contains(forumBaseUrl, "%s") {
		url = fmt.Sprintf(forumBaseUrl, forumSuffix)
	}

	conn := models.CloneConnection()
	defer conn.Close()

	doc, err := goquery.NewDocument(url)
	if err != nil {
		return []*models.Post{}, err
	}

	// find all of the topics
	posts := make([]*models.Post, 0)
	found := 0
	doc.Find("tr[itemtype='http://schema.org/Article']").Each(func(i int, s *goquery.Selection) {

		// we don't want any that are pinned
		pinned := s.Find("span:contains('Pinned')")
		if pinned.Size() == 0 {
			found++
			if found < 5 {
				title := s.Find("span[itemprop='name']").Text()
				url, _ := s.Find("a[itemprop='url']").Attr("href")

				// stop now if we have already saved the topic
				exists, err := models.PostExists(conn, url)
				if err != nil || exists {
					return
				}

				publishedAt, err := getDateFromPost(url)
				if err != nil {
					return
				}

				post := &models.Post{
					Title:       title,
					Source:      "forum",
					Url:         url,
					UniqueId:    url,
					PublishedAt: publishedAt,
				}

				posts = append(posts, post)
			}
		}
	})

	return posts, nil
}

// getDateFromPost returns the post date of a given topic.
func getDateFromPost(url string) (time.Time, error) {
	postDoc, err := goquery.NewDocument(url)
	if err != nil {
		return time.Now(), err
	}

	firstPost := postDoc.Find("#ips_Posts").First()
	timeTitle, _ := firstPost.Find("abbr[itemprop='commentTime']").Attr("title")
	return time.Parse("2006-01-02T15:04:05-07:00", timeTitle)
}
