package updaters

import (
	"fmt"
	"github.com/robmerrell/vtcboard/models"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Network struct{}

var networkBaseUrl = "http://explorer.vertcoin.org/chain/Vertcoin/q"

// Update retrieves VTC netork information from a blockchain api.
func (n *Network) Update() error {
	hashRate, err := getHashRate()
	if err != nil {
		return err
	}

	diff, err := getDifficulty()
	if err != nil {
		return err
	}

	mined, err := getMined()
	if err != nil {
		return err
	}

	blockCount, err := getBlockCount()
	if err != nil {
		return err
	}

	conn := models.CloneConnection()
	defer conn.Close()

	network := &models.Network{
		HashRate:    hashRate,
		Difficulty:  diff,
		Mined:       mined,
		BlockCount:  blockCount,
		GeneratedAt: time.Now().UTC(),
	}
	return network.Insert(conn)
}

// networkQuery queryies the network api at the given url.
func networkQuery(url string) (string, error) {
	resp, err := http.Get(networkBaseUrl + url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	trimmed := strings.TrimSpace(string(body))

	return trimmed, nil
}

// getHashRate gets the current hash rate and converts into a human readable number.
func getHashRate() (string, error) {
	hash, err := networkQuery("/nethash/120/-121/-1")
	if err != nil {
		return "", err
	}

	split := strings.Split(hash, "\n")
	info := strings.TrimSpace(split[len(split)-1])

	nethashSplit := strings.Split(info, ",")
	nethash := nethashSplit[len(nethashSplit)-1]

	converted, err := strconv.ParseFloat(nethash, 64)
	if err != nil {
		return "", err
	}

	mhs := converted / 1000000

	return fmt.Sprintf("%.2f", mhs), nil
}

// getDifficulty gets the current network difficulty
func getDifficulty() (string, error) {
	return networkQuery("/getdifficulty")
}

// getMined gets the total number of mined coins
func getMined() (string, error) {
	mined, err := networkQuery("/totalbc")
	if err != nil {
		return "", err
	}

	converted, err := strconv.ParseFloat(mined, 64)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%.0f", converted), nil
}

// getBlockCount gets the current block count
func getBlockCount() (string, error) {
	return networkQuery("/getblockcount")
}
