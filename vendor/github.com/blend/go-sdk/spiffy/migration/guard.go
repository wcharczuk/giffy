package migration

import (
	"database/sql"

	"github.com/blend/go-sdk/spiffy"
)

// Guard is a control for migration steps.
type Guard func(s *Step, c *spiffy.Connection, tx *sql.Tx) error
