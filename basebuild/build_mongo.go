// +build !sqlite,!mysql,!postgres

package basebuild

import (
	// MongoDB based database service
	_ "github.com/ottemo/foundation/db/mongo"
)
