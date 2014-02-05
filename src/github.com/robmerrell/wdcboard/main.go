package main

import (
	"fmt"
	"github.com/robmerrell/comandante"
	"github.com/robmerrell/wdcboard/cmds"
	"github.com/robmerrell/wdcboard/config"
	"github.com/robmerrell/wdcboard/models"
	"github.com/robmerrell/wdcboard/updaters"
	"os"
)

func main() {
	// get the environment for the config
	appEnv := ""
	env := os.Getenv("WDCBOARD_ENV")
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

	bin := comandante.New("wdcboard", "Worldcoin dashboard")
	bin.IncludeHelp()

	// add indexes to the database
	addIndexes := comandante.NewCommand("index", "Add indexes to the database", cmds.IndexAction)
	addIndexes.Documentation = cmds.IndexDoc
	bin.RegisterCommand(addIndexes)

	// update worldcoin prices
	updateCoinPrices := comandante.NewCommand("update_coin_prices", "Get updated worldcoin prices", cmds.UpdateAction(&updaters.CoinPrice{}))
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

	// update forum posts
	updateForum := comandante.NewCommand("update_forum", "Get new forum topics", cmds.UpdateAction(&updaters.Forum{}))
	updateForum.Documentation = cmds.UpdateForumDoc
	bin.RegisterCommand(updateForum)

	// update reddit stories
	updateReddit := comandante.NewCommand("update_reddit", "Get new /r/worldcoin posts", cmds.UpdateAction(&updaters.Reddit{}))
	updateReddit.Documentation = cmds.UpdateRedditDoc
	bin.RegisterCommand(updateReddit)

	// run web service
	webService := comandante.NewCommand("serve", "Start the WDCBoard web server", cmds.ServeAction)
	webService.Documentation = cmds.ServerDoc
	bin.RegisterCommand(webService)

	if err := bin.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
