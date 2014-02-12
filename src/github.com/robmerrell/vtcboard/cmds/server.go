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

	homeWriter := func(useBtc bool, res http.ResponseWriter) string {
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
		allAverages, err := addLatestPricesToAverages(conn, averages)
		if err != nil {
			webError(err, res)
			return ""
		}

		var graphAverages, graphValueType string
		if useBtc {
			graphValueType = "BTC"
			graphAverages = parseAverages(allAverages, true)
		} else {
			graphValueType = "USD"
			graphAverages = parseAverages(allAverages, false)
		}

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
			"averages":             graphAverages,
			"graphValueType":       graphValueType,
			"showBtcLink":          !useBtc,
			"showUsdLink":          useBtc,
		}

		return mainView.Render(generateTplVars(price, network), valueMap)
	}

	m.Get("/", func(res http.ResponseWriter) string {
		return homeWriter(false, res)
	})

	m.Get("/:graphValue", func(params martini.Params, res http.ResponseWriter) string {
		var useBtc bool
		if params["graphValue"] == "usd" {
			useBtc = false
		} else {
			useBtc = true
		}

		return homeWriter(useBtc, res)
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

	log.Printf("listening on port 4000")
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
func parseAverages(averages []*models.Average, useBtcField bool) string {
	var format string
	if useBtcField {
		format = "[%g, %.8f]"
	} else {
		format = "[%g, %.2f]"
	}

	parsed := ""
	for i, average := range averages {
		if math.IsNaN(average.Cryptsy.Usd) || math.IsNaN(average.Cryptsy.Btc) {
			continue
		}

		timeIndex := float64(average.TimeBlock.Unix()) * 1000.0
		var value float64
		if useBtcField {
			value = average.Cryptsy.Btc
		} else {
			value = average.Cryptsy.Usd
		}

		parsed += fmt.Sprintf(format, timeIndex, value)

		if i < len(averages)-1 {
			parsed += ","
		}
	}

	return parsed
}

// addLatestPriceToAverages appends the latest 10 minutes of price data onto the list of averages
func addLatestPricesToAverages(conn *models.MgoConnection, averages []*models.Average) ([]*models.Average, error) {
	// get times from the last 10 minutes
	baseTime := time.Now().UTC().Truncate(time.Minute * 10)
	beginning := baseTime.Add(time.Minute * -10)
	end := baseTime.Add(time.Minute*-1 + time.Second*59)

	// add the last 10 minutes of pricing data to the end
	prices, err := models.GetPricesBetweenDates(conn, beginning, end)
	if err != nil {
		return []*models.Average{}, err
	}

	allAverages := make([]*models.Average, len(averages), len(averages))
	copy(allAverages, averages)

	for _, price := range prices {
		average := &models.Average{
			TimeBlock: price.GeneratedAt,
			Cryptsy: &models.ExchangeAverage{
				Btc: price.Cryptsy.Btc,
				Usd: price.Cryptsy.Usd,
			},
		}

		allAverages = append(allAverages, average)
	}

	return allAverages, nil
}
