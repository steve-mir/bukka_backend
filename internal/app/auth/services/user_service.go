package services

import (
	"time"
)

// MaxAccountRecoveryDuration is the duration for which an account can be recovered after deletion.
const (
	MaxAccountRecoveryDuration = 30 * 24 * time.Hour // for example, 30 days

)
