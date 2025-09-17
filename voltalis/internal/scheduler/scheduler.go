package scheduler

import "time"

func Run(delay time.Duration, f func() error) error {
	for {
		if err := f(); err != nil {
			return err
		}
		time.Sleep(delay)
	}
}
