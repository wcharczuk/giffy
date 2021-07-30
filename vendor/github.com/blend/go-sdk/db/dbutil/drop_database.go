/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package dbutil

import (
	"context"
	"fmt"

	"github.com/blend/go-sdk/db"
)

// DropDatabase drops a database.
func DropDatabase(ctx context.Context, name string, opts ...db.Option) error {
	conn, err := OpenManagementConnection(opts...)
	if err != nil {
		return err
	}
	if err := CloseAllConnections(ctx, conn, name); err != nil {
		return err
	}
	_, err = conn.ExecContext(ctx, fmt.Sprintf("DROP DATABASE %s ", name))
	return err
}
