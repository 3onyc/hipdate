package hipdate

import (
	_ "github.com/3onyc/hipdate/backends/hipache"
	_ "github.com/3onyc/hipdate/backends/vulcand"

	_ "github.com/3onyc/hipdate/sources/docker"
	_ "github.com/3onyc/hipdate/sources/file"
)
