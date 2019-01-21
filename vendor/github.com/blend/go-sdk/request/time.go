package request

import "time"

func now() time.Time {
	return time.Now().UTC()
}

func since(t time.Time) time.Duration {
	return now().Sub(t)
}
