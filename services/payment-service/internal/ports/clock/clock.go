package clock

import "time"

// Clock abstracts time sourcing
type Clock interface {
	Now() time.Time
}
