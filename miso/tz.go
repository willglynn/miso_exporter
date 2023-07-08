package miso

import (
	"time"
)

// "EST"
var tz *time.Location

func init() {
	tz = time.FixedZone("EST", -5*3600)
}
