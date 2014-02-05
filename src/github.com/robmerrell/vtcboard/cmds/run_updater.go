package cmds

import (
	"github.com/robmerrell/vtcboard/updaters"
)

var UpdateCoinPricesDoc = `
Get updated USD and BTC buy prices from multiple exchange apis.
`

var UpdateNetworkDoc = `
Get updated information about the Vertcoin network.
`

var UpdateRedditDoc = `
Get new posts from /r/vertcoin
`

// UpdateAction returns a function that invokes an updaters Update method
// to be used by comandante.
func UpdateAction(updater updaters.Updater) func() error {
	return func() error {
		return updater.Update()
	}
}
