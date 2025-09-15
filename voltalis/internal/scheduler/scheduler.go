package scheduler

import "time"

func Run(delay time.Duration, f func()) {
	for {
		f()
		time.Sleep(delay)
	}
}
