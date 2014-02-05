package models

import (
	"labix.org/v2/mgo/bson"
	"time"
)

type Network struct {
	HashRate    string    "hashRate"
	Difficulty  string    "difficulty"
	Mined       string    "mined"
	BlockCount  string    "blockCount"
	GeneratedAt time.Time "generatedAt"
}

var networkCollection = "network"

// Insert saves a new WDC network snapshot to the database.
func (n *Network) Insert(conn *MgoConnection) error {
	return conn.DB.C(networkCollection).Insert(n)
}

// GetLatestNetworkSnapshot gets the latest pricing information
func GetLatestNetworkSnapshot(conn *MgoConnection) (*Network, error) {
	var network *Network
	err := conn.DB.C(networkCollection).Find(bson.M{}).Sort("-_id").One(&network)
	return network, err
}
