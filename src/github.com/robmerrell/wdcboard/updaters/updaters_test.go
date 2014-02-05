package updaters

import (
	"fmt"
	"github.com/robmerrell/wdcboard/config"
	"github.com/robmerrell/wdcboard/models"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

// --------------------------------
// Tests for retrieving coin prices
// --------------------------------
type coinPriceSuite struct {
	coinBaseServer *httptest.Server
	cryptsyServer  *httptest.Server

	usdServer *httptest.Server
	badServer *httptest.Server
}

var _ = Suite(&coinPriceSuite{})

func (s *coinPriceSuite) SetUpSuite(c *C) {
	s.coinBaseServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := "{\"btc_to_usd\":\"676.58046\"}"
		fmt.Fprintln(w, res)
	}))

	s.cryptsyServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := "{\"success\": 1, \"return\": {\"markets\": {\"WDC\": {\"recenttrades\": [{\"id\": \"9496223\", \"time\": \"2013-12-25 16:27:42\", \"price\": \"0.00053275\", \"quantity\": \"32.67455654\", \"total\": \"0.01740737\"}]}}}}"
		fmt.Fprintln(w, res)
	}))

	s.badServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "An error occured, I'm not returning valid JSON")
	}))
}

func (s *coinPriceSuite) SetUpTest(c *C) {
	config.LoadConfig("test")
	models.ConnectToDB(config.String("database.host"), config.String("database.db"))
	models.DropCollections()
}

func (s *coinPriceSuite) TearDownSuite(c *C) {
	s.coinBaseServer.Close()
	s.cryptsyServer.Close()
	s.badServer.Close()
}

func replaceUrl(newUrl string, val *string, testFunc func()) {
	oldUrl := *val
	*val = newUrl

	testFunc()

	val = &oldUrl
}

func (s *coinPriceSuite) TestTradePrices(c *C) {
	replaceUrl(s.coinBaseServer.URL, &coinbaseUrl, func() {
		value, _ := coinbaseQuote()
		c.Check(value, Equals, 676.58046)
	})

	replaceUrl(s.badServer.URL, &coinbaseUrl, func() {
		value, err := coinbaseQuote()
		c.Check(value, Equals, 0.0)
		c.Assert(err, NotNil)
	})

	replaceUrl(s.cryptsyServer.URL, &cryptsyUrl, func() {
		value, _ := cryptsyQuote()
		c.Check(value, Equals, 0.00053275)
	})

	replaceUrl(s.badServer.URL, &cryptsyUrl, func() {
		value, err := cryptsyQuote()
		c.Check(value, Equals, 0.0)
		c.Assert(err, NotNil)
	})
}

func (s *coinPriceSuite) TestSavingPrices(c *C) {
	replaceUrl(s.coinBaseServer.URL, &coinbaseUrl, func() {
		replaceUrl(s.cryptsyServer.URL, &cryptsyUrl, func() {
			conn := models.CloneConnection()
			defer conn.Close()

			coinPrice := &CoinPrice{}
			coinPrice.Update()

			var saved models.Price
			conn.DB.C("prices").Find(bson.M{}).One(&saved)

			c.Check(saved.UsdPerBtc, Equals, 676.58046)
			c.Check(saved.Cryptsy.Btc, Equals, 0.00053275)
		})
	})
}

// ---------------------------------
// Tests for retrieving network info
// ---------------------------------
type networkSuite struct {
	hashRateServer   *httptest.Server
	difficultyServer *httptest.Server
	minedServer      *httptest.Server
	blockcountServer *httptest.Server
}

var _ = Suite(&networkSuite{})

func (s *networkSuite) SetUpSuite(c *C) {
	s.hashRateServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := "6792543827"
		fmt.Fprintln(w, res)
	}))

	s.difficultyServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := "42.177"
		fmt.Fprintln(w, res)
	}))

	s.minedServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := "37755394.06249219"
		fmt.Fprintln(w, res)
	}))

	s.blockcountServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := "915281"
		fmt.Fprintln(w, res)
	}))
}

func (s *networkSuite) SetUpTest(c *C) {
	config.LoadConfig("test")
	models.ConnectToDB(config.String("database.host"), config.String("database.db"))
	models.DropCollections()
}

func (s *networkSuite) TearDownSuite(c *C) {
	s.hashRateServer.Close()
	s.difficultyServer.Close()
	s.minedServer.Close()
	s.blockcountServer.Close()
}

func (s *networkSuite) TestNetworkCalls(c *C) {
	replaceUrl(s.hashRateServer.URL, &networkBaseUrl, func() {
		value, _ := getHashRate()
		c.Check(value, Equals, "6792.54")
	})

	replaceUrl(s.difficultyServer.URL, &networkBaseUrl, func() {
		value, _ := getDifficulty()
		c.Check(value, Equals, "42.177")
	})

	replaceUrl(s.minedServer.URL, &networkBaseUrl, func() {
		value, _ := getMined()
		c.Check(value, Equals, "37755394")
	})

	replaceUrl(s.blockcountServer.URL, &networkBaseUrl, func() {
		value, _ := getBlockCount()
		c.Check(value, Equals, "915281")
	})
}

// --------------------------
// Tests for retrieving posts
// --------------------------
type postSuite struct {
	redditServer1    *httptest.Server
	redditServer2    *httptest.Server
	forumListServer  *httptest.Server
	forumTopicServer *httptest.Server
}

var _ = Suite(&postSuite{})

func (s *postSuite) SetUpSuite(c *C) {
	s.redditServer1 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := `<?xml version="1.0" encoding="UTF-8"?> <rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:media="http://search.yahoo.com/mrss/" xmlns:atom="http://www.w3.org/2005/Atom"> <channel> <title>worldcoin cryptocurrency</title> <link>http://www.reddit.com/r/worldcoin/</link> <description>worldcoin cryptocurrency</description> <image> <url>http://www.reddit.com/reddit.com.header.png</url> <title>worldcoin cryptocurrency</title> <link>http://www.reddit.com/r/worldcoin/</link> </image> <atom:link rel="self" href="http://www.reddit.com/r/worldcoin/.rss" type="application/rss+xml" /> <item> <title>two worldcoins?</title> <link>http://www.reddit.com/r/worldcoin/comments/1ujf68/two_worldcoins/</link> <guid isPermaLink="true">http://www.reddit.com/r/worldcoin/comments/1ujf68/two_worldcoins/</guid> <pubDate>Mon, 06 Jan 2014 14:29:39 +0000</pubDate> <description>&lt;!-- SC_OFF --&gt;&lt;div class=&#34;md&#34;&gt;&lt;p&gt;Hi,&lt;/p&gt; &lt;p&gt;I like the look of WDC and have purchased a few and intend to get myself some more but Im a little confused with this site: &lt;a href=&#34;http://www.coinworld.org/&#34;&gt;http://www.coinworld.org/&lt;/a&gt;&lt;/p&gt; &lt;p&gt;If I&amp;#39;m not mistaken this is another worldcoin?! That is surely not a good thing!&lt;/p&gt; &lt;p&gt;Cheers&lt;/p&gt; &lt;/div&gt;&lt;!-- SC_ON --&gt; submitted by &lt;a href=&#34;http://www.reddit.com/user/smi2ler&#34;&gt; smi2ler &lt;/a&gt; &lt;br/&gt; &lt;a href=&#34;http://www.reddit.com/r/worldcoin/comments/1ujf68/two_worldcoins/&#34;&gt;[link]&lt;/a&gt; &lt;a href="http://www.reddit.com/r/worldcoin/comments/1ujf68/two_worldcoins/"&gt;[8 comments]&lt;/a&gt;</description> </item> <item> <title>Whats a better name than Scharmbeck?</title> <link>http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/</link> <guid isPermaLink="true">http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/</guid> <pubDate>Mon, 06 Jan 2014 10:32:36 +0000</pubDate> <description>&lt;!-- SC_OFF --&gt;&lt;div class=&#34;md&#34;&gt;&lt;p&gt;This has been brought up before but I feel it needs more attention from the foundation. The name &amp;quot;Scharmbeck&amp;quot;, while giving off the vibe of a historical German bank is not catchy, memorable or modern. What do you redditors feel would be a better name? &lt;/p&gt; &lt;p&gt;Special thanks to the foundation for all the great work they have done so far. &lt;/p&gt; &lt;/div&gt;&lt;!-- SC_ON --&gt; submitted by &lt;a href=&#34;http://www.reddit.com/user/kanada_kid&#34;&gt; kanada_kid &lt;/a&gt; &lt;br/&gt; &lt;a href=&#34;http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/&#34;&gt;[link]&lt;/a&gt; &lt;a href="http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/"&gt;[13 comments]&lt;/a&gt;</description> </item> </channel> </rss>`
		fmt.Fprintln(w, res)
	}))

	s.redditServer2 = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res := `<?xml version="1.0" encoding="UTF-8"?> <rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:media="http://search.yahoo.com/mrss/" xmlns:atom="http://www.w3.org/2005/Atom"> <channel> <title>worldcoin cryptocurrency</title> <link>http://www.reddit.com/r/worldcoin/</link> <description>worldcoin cryptocurrency</description> <image> <url>http://www.reddit.com/reddit.com.header.png</url> <title>worldcoin cryptocurrency</title> <link>http://www.reddit.com/r/worldcoin/</link> </image> <atom:link rel="self" href="http://www.reddit.com/r/worldcoin/.rss" type="application/rss+xml" /> <item> <title>two worldcoins?</title> <link>http://www.reddit.com/r/worldcoin/comments/1ujf68/two_worldcoins/</link> <guid isPermaLink="true">firstguid</guid> <pubDate>Mon, 06 Jan 2014 14:29:39 +0000</pubDate> <description>&lt;!-- SC_OFF --&gt;&lt;div class=&#34;md&#34;&gt;&lt;p&gt;Hi,&lt;/p&gt; &lt;p&gt;I like the look of WDC and have purchased a few and intend to get myself some more but Im a little confused with this site: &lt;a href=&#34;http://www.coinworld.org/&#34;&gt;http://www.coinworld.org/&lt;/a&gt;&lt;/p&gt; &lt;p&gt;If I&amp;#39;m not mistaken this is another worldcoin?! That is surely not a good thing!&lt;/p&gt; &lt;p&gt;Cheers&lt;/p&gt; &lt;/div&gt;&lt;!-- SC_ON --&gt; submitted by &lt;a href=&#34;http://www.reddit.com/user/smi2ler&#34;&gt; smi2ler &lt;/a&gt; &lt;br/&gt; &lt;a href=&#34;http://www.reddit.com/r/worldcoin/comments/1ujf68/two_worldcoins/&#34;&gt;[link]&lt;/a&gt; &lt;a href="http://www.reddit.com/r/worldcoin/comments/1ujf68/two_worldcoins/"&gt;[8 comments]&lt;/a&gt;</description> </item> <item> <title>Whats a better name than Scharmbeck?</title> <link>http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/</link> <guid isPermaLink="true">secondguid</guid> <pubDate>Mon, 06 Jan 2014 10:32:36 +0000</pubDate> <description>&lt;!-- SC_OFF --&gt;&lt;div class=&#34;md&#34;&gt;&lt;p&gt;This has been brought up before but I feel it needs more attention from the foundation. The name &amp;quot;Scharmbeck&amp;quot;, while giving off the vibe of a historical German bank is not catchy, memorable or modern. What do you redditors feel would be a better name? &lt;/p&gt; &lt;p&gt;Special thanks to the foundation for all the great work they have done so far. &lt;/p&gt; &lt;/div&gt;&lt;!-- SC_ON --&gt; submitted by &lt;a href=&#34;http://www.reddit.com/user/kanada_kid&#34;&gt; kanada_kid &lt;/a&gt; &lt;br/&gt; &lt;a href=&#34;http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/&#34;&gt;[link]&lt;/a&gt; &lt;a href="http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/"&gt;[13 comments]&lt;/a&gt;</description> </item> </channel> </rss>`
		fmt.Fprintln(w, res)
	}))

	s.forumListServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		posts := `
		<html>
		<body>
		<table>
		<tr itemtype='http://schema.org/Article'>
			<td>
				<span>Pinned</span>
			</td>
		</tr>
		<tr itemtype='http://schema.org/Article'>
			<td>
				<span itemprop='name'>Title1</span>
				<a itemprop='url' href='%s'>url</a>
			</td>
		</tr>
		<tr itemtype='http://schema.org/Article'>
			<td>
				<span itemprop='name'>Title2</span>
				<a itemprop='url' href='%s'>url2</a>
			</td>
		</tr>	
		</table>	
		</body>
		</html>
		`

		fmt.Fprintln(w, fmt.Sprintf(posts, s.forumTopicServer.URL, s.forumTopicServer.URL))
	}))

	s.forumTopicServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		topic := `
		<html>
		<body>
		<div id='ips_Posts'>
			<div><abbr itemprop='commentTime' title='2014-01-13T17:33:05+00:00'></abbr></div>
			<div><abbr itemprop='commentTime' title='2014-01-14T17:33:05+00:00'></abbr></div>
		</div>
		</body>
		</html>
		`
		fmt.Fprintln(w, topic)
	}))
}

func (s *postSuite) SetUpTest(c *C) {
	config.LoadConfig("test")
	models.ConnectToDB(config.String("database.host"), config.String("database.db"))
	models.DropCollections()
}

func (s *postSuite) TearDownSuite(c *C) {
	s.redditServer1.Close()
	s.redditServer2.Close()
	s.forumListServer.Close()
	s.forumTopicServer.Close()
}

func (s *postSuite) TestGettingReditPosts(c *C) {
	conn := models.CloneConnection()
	defer conn.Close()

	replaceUrl(s.redditServer1.URL, &subredditUrl, func() {
		p1 := &models.Post{Title: "test title", Url: "test url", Source: "reddit", UniqueId: "http://www.reddit.com/r/worldcoin/comments/1uj486/whats_a_better_name_than_scharmbeck/"}
		p1.Insert(conn)

		posts, _ := getNewRedditPosts()

		c.Check(len(posts), Equals, 1)
		c.Check(posts[0].Title, Equals, "two worldcoins?")
	})
}

func (s *postSuite) TestUpdatingReddit(c *C) {
	conn := models.CloneConnection()
	defer conn.Close()

	replaceUrl(s.redditServer2.URL, &subredditUrl, func() {
		r := &Reddit{}
		r.Update()

		var results []*models.Post
		conn.DB.C("posts").Find(bson.M{}).All(&results)

		c.Check(len(results), Equals, 2)
	})
}

func (s *postSuite) TestUpdateForum(c *C) {
	replaceUrl(s.forumListServer.URL, &forumBaseUrl, func() {
		posts, _ := getNewTopicsFromForum("testing")

		c.Check(len(posts), Equals, 2)
		c.Check(posts[0].Title, Equals, "Title1")
		c.Check(posts[1].Title, Equals, "Title2")

		parsedTime, _ := time.Parse("2006-01-02T15:04:05-07:00", "2014-01-13T17:33:05+00:00")
		c.Check(posts[0].PublishedAt.Format(time.RFC3339), Equals, parsedTime.Format(time.RFC3339))
	})
}
