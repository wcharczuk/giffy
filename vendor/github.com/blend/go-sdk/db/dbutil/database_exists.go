/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package dbutil

import (
	"context"

	"github.com/blend/go-sdk/db"
)

// DatabaseExists returns if a database exists or not.
func DatabaseExists(ctx context.Context, name string, opts ...db.Option) (bool, error) {
	conn, err := OpenManagementConnection(opts...)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	return conn.QueryContext(ctx, "SELECT 1 FROM pg_database WHERE datname = $1", name).Any()
}
