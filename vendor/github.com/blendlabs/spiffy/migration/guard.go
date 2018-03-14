package migration

import (
	"database/sql"

	"github.com/blendlabs/spiffy"
)

// Guard is a control for migration steps.
type Guard func(s *Step, c *spiffy.Connection, tx *sql.Tx) error
