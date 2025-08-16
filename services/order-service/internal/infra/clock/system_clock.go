package clock

import "time"

// SystemClock returns the system time
type SystemClock struct{}

func NewSystemClock() *SystemClock { return &SystemClock{} }

func (SystemClock) Now() time.Time { return time.Now() }
