/*

Copyright (c) 2021 - Present. Blend Labs, Inc. All rights reserved
Use of this source code is governed by a MIT license that can be found in the LICENSE file.

*/

package db

import (
	"context"
)

// StatementInterceptor is an interceptor for statements.
type StatementInterceptor func(ctx context.Context, label, statement string) (string, error)
