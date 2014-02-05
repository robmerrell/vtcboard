package cmds

import (
	"github.com/robmerrell/vtcboard/models"
	"time"
)

var PricingRollupDoc = `
Aggregate the pricing information to an interval set in the config file. This is
for grouping together data to be showed on a chart.
`

func PricingRollupAction() error {
	conn := models.CloneConnection()
	defer conn.Close()

	// get all prices from the last 10 minutes
	baseTime := time.Now().UTC().Truncate(time.Minute * 10)
	beginning := baseTime.Add(time.Minute * -10)
	end := baseTime.Add(time.Minute*-1 + time.Second*59)

	_, err := models.GenerateAverage(conn, beginning, end)
	return err
}
