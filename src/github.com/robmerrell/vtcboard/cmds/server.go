package cmds

import (
	"errors"
	"fmt"
	"github.com/codegangsta/martini"
	"github.com/hoisie/mustache"
	"github.com/robmerrell/vtcboard/lib"
	"github.com/robmerrell/vtcboard/models"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"
)

var ServerDoc = `
Starts the WDCBoard webserver.
`

func webError(err error, res http.ResponseWriter) {
	log.Println(err)
	http.Error(res, "There was an error, try again later", 500)
}

func ServeAction() error {
	m := martini.Classic()
	m.Use(martini.Static("resources/public"))

	mainView, err := mustache.ParseFile("resources/views/main.html.mustache")
	if err != nil {
		panic(err)
	}

	m.Get("/", func(res http.ResponseWriter) string {
		conn := models.CloneConnection()
		defer conn.Close()

		// get the latest pricing data
		price, err := models.GetLatestPrice(conn)
		if err != nil {
			webError(err, res)
			return ""
		}

		// get data for the graph
		averages, err := models.GetAverages(conn, 24)
		if err != nil {
			webError(err, res)
			return ""
		}
		parsedAverages := parseAverages(averages)

		// get the forum posts
		forum, err := models.GetLatestPosts(conn, "forum", 8)
		if err != nil {
			webError(err, res)
			return ""
		}

		// /r/vertcoin posts
		redditVertcoin, err := models.GetLatestPosts(conn, "/r/vertcoin", 8)
		if err != nil {
			webError(err, res)
			return ""
		}

		// /r/vertmarket posts
		redditVertmarket, err := models.GetLatestPosts(conn, "/r/vertmarket", 8)
		if err != nil {
			webError(err, res)
			return ""
		}

		// /r/vertmining posts
		redditVertcoinMining, err := models.GetLatestPosts(conn, "/r/vertcoinmining", 8)
		if err != nil {
			webError(err, res)
			return ""
		}

		// get the mining information
		network, err := models.GetLatestNetworkSnapshot(conn)
		if err != nil {
			webError(err, res)
			return ""
		}

		// generate the HTML
		valueMap := map[string]interface{}{
			"redditVertcoin":       redditVertcoin,
			"redditVertmarket":     redditVertmarket,
			"redditVertcoinMining": redditVertcoinMining,
			"forum":                forum,
			"averages":             parsedAverages,
		}
		return mainView.Render(generateTplVars(price, network), valueMap)
	})

	// returns basic information about the state of the service. If any hardcoded checks fail
	// the message is returned with a 500 status. We can then use pingdom or another service
	// to alert when data integrity may be off.
	m.Get("/health", func(res http.ResponseWriter) string {
		conn := models.CloneConnection()
		defer conn.Close()

		twoHoursAgo := time.Now().Add(time.Hour * -2).Unix()

		// make sure the price has been updated in the last 2 hours
		price, err := models.GetLatestPrice(conn)
		if err != nil {
			webError(errors.New("Error getting latest price"), res)
			return ""
		}

		if price.GeneratedAt.Unix() < twoHoursAgo {
			webError(errors.New("The latest price is old"), res)
			return ""
		}

		// make sure the network has been updated in the last two hours
		network, err := models.GetLatestNetworkSnapshot(conn)
		if err != nil {
			webError(errors.New("Error getting latest network snapshot"), res)
			return ""
		}

		if network.GeneratedAt.Unix() < twoHoursAgo {
			webError(errors.New("The latest network snapshot is old"), res)
			return ""
		}

		return "ok"
	})

	http.ListenAndServe(":4000", m)

	return nil
}

// generateTplVars generates a map to pass into the template
func generateTplVars(price *models.Price, network *models.Network) map[string]string {
	// apply the necessary style for the percent change box
	changeStyle := "percent-change-stat-up"
	if price.Cryptsy.PercentChange != "" && string(price.Cryptsy.PercentChange[0]) == "-" {
		changeStyle = "percent-change-stat-down"
	}

	percentChange := "100"
	if price.Cryptsy.PercentChange != "" {
		percentChange = price.Cryptsy.PercentChange
	}

	// marketcap
	minedNum, _ := strconv.Atoi(network.Mined)
	marketCap := float64(minedNum) * price.Cryptsy.Usd

	// coins left to be mined
	remainingCoins := 84000000 - minedNum

	vars := map[string]string{
		"usd":         lib.RenderFloat("", price.Cryptsy.Usd),
		"btc":         strconv.FormatFloat(price.Cryptsy.Btc, 'f', 8, 64),
		"marketCap":   lib.RenderInteger("", int(marketCap)),
		"change":      percentChange,
		"changeStyle": changeStyle,

		"hashRate":   lib.RenderFloatFromString("", network.HashRate),
		"difficulty": lib.RenderFloatFromString("", network.Difficulty),
		"mined":      lib.RenderIntegerFromString("", network.Mined),
		"remaining":  lib.RenderInteger("", remainingCoins),
	}

	return vars
}

// parseAverages takes a slice of averages and returns a string representation for flot to graph
func parseAverages(averages []*models.Average) string {
	parsed := ""
	for i, average := range averages {
		if math.IsNaN(average.Cryptsy.Usd) {
			continue
		}

		timeIndex := float64(average.TimeBlock.Unix()) * 1000.0
		parsed += fmt.Sprintf("[%g, %.2f]", timeIndex, average.Cryptsy.Usd)

		if i < len(averages) {
			parsed += ","
		}
	}

	return parsed
}
