package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/db"
)

// Guard is a control for migration steps.
type Guard func(s *Step, c *db.Connection, tx *sql.Tx) error
