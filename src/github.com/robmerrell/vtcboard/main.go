package main

import (
	"fmt"
	"github.com/robmerrell/comandante"
	"github.com/robmerrell/vtcboard/cmds"
	"github.com/robmerrell/vtcboard/config"
	"github.com/robmerrell/vtcboard/models"
	"github.com/robmerrell/vtcboard/updaters"
	"os"
)

func main() {
	// get the environment for the config
	appEnv := ""
	env := os.Getenv("VTCBOARD_ENV")
	switch env {
	case "dev", "test", "prod":
		appEnv = env
	default:
		appEnv = "dev"
	}

	config.LoadConfig(appEnv)

	// make the package level database connection
	if err := models.ConnectToDB(config.String("database.host"), config.String("database.db")); err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to the database")
		os.Exit(1)
	}
	defer models.CloseDB()

	bin := comandante.New("vtcboard", "Vertcoin dashboard")
	bin.IncludeHelp()

	// add indexes to the database
	addIndexes := comandante.NewCommand("index", "Add indexes to the database", cmds.IndexAction)
	addIndexes.Documentation = cmds.IndexDoc
	bin.RegisterCommand(addIndexes)

	// update vertcoin prices
	updateCoinPrices := comandante.NewCommand("update_coin_prices", "Get updated vertcoin prices", cmds.UpdateAction(&updaters.CoinPrice{}))
	updateCoinPrices.Documentation = cmds.UpdateCoinPricesDoc
	bin.RegisterCommand(updateCoinPrices)

	// pricing rollup for the graph
	rollupPricing := comandante.NewCommand("pricing_rollup", "Aggregate pricing information", cmds.PricingRollupAction)
	rollupPricing.Documentation = cmds.PricingRollupDoc
	bin.RegisterCommand(rollupPricing)

	// update network info
	updateNetwork := comandante.NewCommand("update_network", "Get updated network information", cmds.UpdateAction(&updaters.Network{}))
	updateNetwork.Documentation = cmds.UpdateCoinPricesDoc
	bin.RegisterCommand(updateNetwork)

	// update reddit stories
	updateReddit := comandante.NewCommand("update_reddit", "Get new /r/vertcoin posts", cmds.UpdateAction(&updaters.Reddit{}))
	updateReddit.Documentation = cmds.UpdateRedditDoc
	bin.RegisterCommand(updateReddit)

	// run web service
	webService := comandante.NewCommand("serve", "Start the VTCBoard web server", cmds.ServeAction)
	webService.Documentation = cmds.ServerDoc
	bin.RegisterCommand(webService)

	if err := bin.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
