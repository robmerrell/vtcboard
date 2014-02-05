package models

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type Post struct {
	Title       string    "title"
	Source      string    "source"
	Url         string    "url"
	UniqueId    string    "uniqueId"
	PublishedAt time.Time "publishedAt"
}

var postCollection = "posts"

// PostExists checks if a post at a given uniqueId already exists.
func PostExists(conn *MgoConnection, id string) (bool, error) {
	var result Post
	err := conn.DB.C(postCollection).Find(bson.M{"uniqueId": id}).One(&result)

	if err != nil {
		if err == mgo.ErrNotFound {
			return false, nil
		} else {
			return false, err
		}
	}

	return true, nil
}

// GetLatestPosts returns the latests posts of the given source.
func GetLatestPosts(conn *MgoConnection, source string, limit int) ([]*Post, error) {
	var posts []*Post
	err := conn.DB.C(postCollection).Find(bson.M{"source": source}).Sort("-publishedAt").Limit(limit).All(&posts)
	return posts, err
}

// Insert saves a new post.
func (p *Post) Insert(conn *MgoConnection) error {
	return conn.DB.C(postCollection).Insert(p)
}
