package scheduler

import "time"

func Run(f func()) {
	for {
		f()
		time.Sleep(15 * time.Second)
	}
}
