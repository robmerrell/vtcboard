package updaters

import (
	"encoding/json"
	"github.com/robmerrell/vtcboard/models"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

var coinbaseUrl = "https://coinbase.com/api/v1/currencies/exchange_rates"
var cryptsyUrl = "http://pubapi.cryptsy.com/api.php?method=singlemarketdata&marketid=14"

type CoinPrice struct{}

// Update retrieves WDC buy prices in both USD and BTC and saves
// the prices to the database.
func (c *CoinPrice) Update() error {
	usd, err := coinbaseQuote()
	if err != nil {
		return err
	}

	cryptsyBtc, err := cryptsyQuote()
	if err != nil {
		return err
	}

	conn := models.CloneConnection()
	defer conn.Close()

	price := &models.Price{
		UsdPerBtc:   usd,
		Cryptsy:     &models.ExchangePrice{Btc: cryptsyBtc, Usd: usd * cryptsyBtc},
		GeneratedAt: time.Now().UTC().Truncate(time.Minute),
	}
	if err := price.SetPercentChange(conn); err != nil {
		return err
	}

	return price.Insert(conn)
}

// coinbaseQuote gets the current USD value for 1 BTC from coinbase.
func coinbaseQuote() (float64, error) {
	resp, err := http.Get(coinbaseUrl)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var value struct {
		Btc string `json:"btc_to_usd"`
	}
	if err := json.Unmarshal(body, &value); err != nil {
		return 0.0, err
	}

	return strconv.ParseFloat(value.Btc, 64)
}

// cryptsyQuote gets the current WDC/BTC value from cryptsy.
func cryptsyQuote() (float64, error) {
	resp, err := http.Get(cryptsyUrl)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	var value struct {
		Return struct {
			Markets struct {
				WDC struct {
					RecentTrades []map[string]string
				}
			}
		}
	}
	if err := json.Unmarshal(body, &value); err != nil {
		return 0.0, err
	}

	return strconv.ParseFloat(value.Return.Markets.WDC.RecentTrades[0]["price"], 64)
}
