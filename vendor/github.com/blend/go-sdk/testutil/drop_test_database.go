/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package testutil

import (
	"context"
	"fmt"

	"github.com/blend/go-sdk/db"

	"github.com/blend/go-sdk/db/dbutil"
)

// DropTestDatabase drops a database.
func DropTestDatabase(ctx context.Context, conn *db.Connection, opts ...db.Option) error {
	config, err := conn.Config.Reparse()
	if err != nil {
		return err
	}

	mgmt, err := dbutil.OpenManagementConnection(opts...)
	if err != nil {
		return err
	}
	_, err = mgmt.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s ", config.Database))
	return err
}
