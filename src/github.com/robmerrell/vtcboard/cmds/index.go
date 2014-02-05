package cmds

import (
	"github.com/robmerrell/vtcboard/models"
)

var IndexDoc = `
Adds needed indexes to the database.
`

// IndexAction is the function invoked by the index command.
func IndexAction() error {
	conn := models.CloneConnection()
	defer conn.Close()
	return models.Index(conn)
}
