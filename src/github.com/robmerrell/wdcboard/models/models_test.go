package models

import (
	"github.com/robmerrell/wdcboard/config"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

// -----------
// Price model
// -----------
type priceSuite struct{}

var _ = Suite(&priceSuite{})

func (s *priceSuite) SetUpTest(c *C) {
	config.LoadConfig("test")
	ConnectToDB(config.String("database.host"), config.String("database.db"))
	DropCollections()
}

func (s *priceSuite) TestInserting(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	p := &Price{
		UsdPerBtc: 100.0,
		Cryptsy: &ExchangePrice{
			Btc: 0.3456,
		},
	}
	p.Insert(conn)

	var newp Price
	conn.DB.C(priceCollection).Find(bson.M{"usdperbtc": 100.0}).One(&newp)

	c.Check(newp.Cryptsy.Btc, Equals, 0.3456)
}

func (s *priceSuite) TestGettingPricesBetweenDates(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	baseTime := time.Now().UTC().Truncate(time.Minute)

	p1 := &Price{UsdPerBtc: 99, GeneratedAt: baseTime}
	p1.Insert(conn)

	d, _ := time.ParseDuration("-24h")
	p2 := &Price{UsdPerBtc: 98, GeneratedAt: baseTime.Add(d)}
	p2.Insert(conn)

	d, _ = time.ParseDuration("24h")
	p3 := &Price{UsdPerBtc: 100, GeneratedAt: baseTime.Add(d)}
	p3.Insert(conn)

	prices, _ := GetPricesBetweenDates(conn, baseTime, p3.GeneratedAt)

	c.Check(len(prices), Equals, 2)
	c.Check(prices[0].UsdPerBtc, Equals, float64(99))
	c.Check(prices[1].UsdPerBtc, Equals, float64(100))
}

func (s *priceSuite) TestSettingPercentChange(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	yesterdayDuration, _ := time.ParseDuration("-25h")

	p1 := &Price{
		UsdPerBtc: 100.0,
		Cryptsy: &ExchangePrice{
			Btc: 0.3456,
			Usd: 1.0,
		},
		GeneratedAt: time.Now().UTC().Add(yesterdayDuration),
	}
	p1.Insert(conn)

	p2 := &Price{
		UsdPerBtc: 100.0,
		Cryptsy: &ExchangePrice{
			Btc: 0.55,
			Usd: 1.45,
		},
		GeneratedAt: time.Now().UTC(),
	}
	p2.SetPercentChange(conn)

	c.Check(p2.Cryptsy.PercentChange, Equals, "45.00")
	c.Check(p2.ChangeComparison, Equals, p1.Id)
}

func (s *priceSuite) TestPercentChange(c *C) {
	c.Check(percentChange(1, 2), Equals, "100.00")
	c.Check(percentChange(1, 5), Equals, "400.00")
	c.Check(percentChange(3, 1.63), Equals, "-45.67")
	c.Check(percentChange(0.456, 0.457), Equals, "0.22")
}

// -------------
// Average model
// -------------
type averageSuite struct{}

var _ = Suite(&averageSuite{})

func (s *averageSuite) SetUpTest(c *C) {
	config.LoadConfig("test")
	ConnectToDB(config.String("database.host"), config.String("database.db"))
	DropCollections()
}

func (s *averageSuite) TestInserting(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	t := time.Now().UTC()
	a := &Average{TimeBlock: t, Cryptsy: &ExchangeAverage{Btc: 0.5}}
	a.Insert(conn)

	var avg Average
	conn.DB.C(averageCollection).Find(bson.M{"timeBlock": t}).One(&avg)

	c.Check(avg.Cryptsy.Btc, Equals, 0.5)
}

func (s *averageSuite) TestGeneratingAverages(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	baseTime := time.Now().UTC().Truncate(time.Minute * 10)
	beginning := baseTime.Add(time.Minute * -10)
	end := baseTime.Add(time.Minute*-1 + time.Second*59)

	p1 := &Price{
		UsdPerBtc: 100.0,
		Cryptsy: &ExchangePrice{
			Btc: 1.0,
			Usd: 98.0,
		},
		GeneratedAt: beginning,
	}
	p1.Insert(conn)

	p2 := &Price{
		UsdPerBtc: 100.0,
		Cryptsy: &ExchangePrice{
			Btc: 3.0,
			Usd: 100.0,
		},
		GeneratedAt: end,
	}
	p2.Insert(conn)

	avg, _ := GenerateAverage(conn, beginning, end)

	c.Check(avg.Cryptsy.Usd, Equals, float64(99))
	c.Check(avg.Cryptsy.Btc, Equals, float64(2))
}

func (s *averageSuite) TestGettingAverages(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	baseTime := time.Now().UTC().Truncate(time.Minute * 10)
	beginning := baseTime.Add(time.Minute * -10)
	end := baseTime.Add(time.Minute*-1 + time.Second*59)

	p1 := &Price{
		UsdPerBtc: 100.0,
		Cryptsy: &ExchangePrice{
			Btc: 1.0,
			Usd: 98.0,
		},
		GeneratedAt: beginning,
	}
	p1.Insert(conn)

	p2 := &Price{
		UsdPerBtc: 100.0,
		Cryptsy: &ExchangePrice{
			Btc: 3.0,
			Usd: 100.0,
		},
		GeneratedAt: end,
	}
	p2.Insert(conn)

	GenerateAverage(conn, beginning, end)
	averages, _ := GetAverages(conn, 10)

	c.Check(len(averages), Equals, 1)
	c.Check(averages[0].Cryptsy.Usd, Equals, float64(99))
}

// -------------
// Network model
// -------------
type networkSuite struct{}

var _ = Suite(&networkSuite{})

func (s *networkSuite) SetUpTest(c *C) {
	config.LoadConfig("test")
	ConnectToDB(config.String("database.host"), config.String("database.db"))
	DropCollections()
}

func (s *networkSuite) TestInserting(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	n := &Network{HashRate: "100.0", Mined: "1234"}
	n.Insert(conn)

	var info Network
	conn.DB.C(networkCollection).Find(bson.M{"hashRate": "100.0"}).One(&info)

	c.Check(info.Mined, Equals, "1234")
}

// -----------
// Posts model
// -----------
type postSuite struct{}

var _ = Suite(&postSuite{})

func (s *postSuite) SetUpTest(c *C) {
	config.LoadConfig("test")
	ConnectToDB(config.String("database.host"), config.String("database.db"))
	DropCollections()
}

func (s *postSuite) TestInserting(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	p := &Post{Title: "test title", Url: "test url"}
	p.Insert(conn)

	var info Post
	conn.DB.C(postCollection).Find(bson.M{"title": "test title"}).One(&info)

	c.Check(info.Url, Equals, "test url")
}

func (s *postSuite) TestPostExists(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	p := &Post{Title: "test title", Url: "test url", UniqueId: "test id"}
	p.Insert(conn)

	exists, _ := PostExists(conn, "test id")
	c.Check(exists, Equals, true)
}

func (s *postSuite) TestGettingLatestPosts(c *C) {
	conn := CloneConnection()
	defer conn.Close()

	t1 := time.Now().UTC().Add(time.Hour * -10)
	t2 := time.Now().UTC()
	t3 := time.Now().UTC().Add(time.Hour * -100)

	p1 := &Post{Title: "test title", Url: "test url", Source: "reddit", PublishedAt: t1}
	p1.Insert(conn)

	p2 := &Post{Title: "test title2", Url: "test url2", Source: "reddit", PublishedAt: t2}
	p2.Insert(conn)

	p3 := &Post{Title: "test title3", Url: "test url3", Source: "reddit", PublishedAt: t3}
	p3.Insert(conn)

	p4 := &Post{Title: "test different source", Url: "test url2", Source: "somethingelse"}
	p4.Insert(conn)

	posts, _ := GetLatestPosts(conn, "reddit", 2)

	c.Check(len(posts), Equals, 2)
	c.Check(posts[0].Title, Equals, "test title2")
	c.Check(posts[1].Title, Equals, "test title")
}
