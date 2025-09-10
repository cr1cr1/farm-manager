package pages

import "time"

func YearNow() int {
	return time.Now().UTC().Year()
}
