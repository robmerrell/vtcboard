package models

import (
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"time"
)

type ExchangePrice struct {
	Btc           float64 "btc"
	Usd           float64 "usd"
	PercentChange string  "percentChange"
}

type Price struct {
	Id               bson.ObjectId  "_id,omitempty"
	UsdPerBtc        float64        "usdperbtc"
	Cryptsy          *ExchangePrice "cryptsy"
	GeneratedAt      time.Time      "generatedAt"
	ChangeComparison bson.ObjectId  "changeComparison,omitempty"
}

var priceCollection = "prices"

// GetLatestPrice gets the latest pricing information
func GetLatestPrice(conn *MgoConnection) (*Price, error) {
	var price *Price
	err := conn.DB.C(priceCollection).Find(bson.M{}).Sort("-_id").One(&price)
	return price, err
}

// GetPricesBetweenDates retrieves all of the prices between two dates, inclusively.
func GetPricesBetweenDates(conn *MgoConnection, beginning, end time.Time) ([]*Price, error) {
	var prices []*Price
	err := conn.DB.C(priceCollection).Find(bson.M{"generatedAt": bson.M{"$gte": beginning, "$lte": end}}).Sort("_id").All(&prices)
	return prices, err
}

// Insert saves a new WDC price point to the database.
func (p *Price) Insert(conn *MgoConnection) error {
	p.Id = bson.NewObjectId()
	return conn.DB.C(priceCollection).Insert(p)
}

// SetPercentChange adds the percent change from the last 24 hours for all exchanges.
func (p *Price) SetPercentChange(conn *MgoConnection) error {
	yesterdayDuration, _ := time.ParseDuration("-24h")
	previousTime := p.GeneratedAt.Add(yesterdayDuration)

	// find the record closest to 24 hours as possible
	var oldPrice Price
	if err := conn.DB.C(priceCollection).Find(bson.M{"generatedAt": bson.M{"$lte": previousTime}}).Sort("-_id").One(&oldPrice); err != nil {
		if err == mgo.ErrNotFound {
			return nil
		} else {
			return err
		}
	}

	p.Cryptsy.PercentChange = percentChange(oldPrice.Cryptsy.Usd, p.Cryptsy.Usd)
	p.ChangeComparison = oldPrice.Id

	return nil
}

// percentChange calculates the percent change between to BTC values.
func percentChange(oldUsd, newUsd float64) string {
	var change float64
	if oldUsd == 0.0 {
		change = 100.0
	} else {
		diff := (newUsd - oldUsd)
		change = (diff / oldUsd) * 100
	}

	return fmt.Sprintf("%.2f", change)
}
