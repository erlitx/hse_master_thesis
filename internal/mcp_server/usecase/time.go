package usecase

import "time"

func (u *UseCase) NowISO() string {
	// RFC3339Nano makes round-trips easier for hosts
	return u.clock.Now().UTC().Format(time.RFC3339Nano)
}
