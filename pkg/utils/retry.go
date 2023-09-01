package utils

import (
	"errors"
	"time"
)

func Retry(amount int, dur time.Duration, x func() error) error {
	if amount <= 0 {
		return errors.New("amount must be larger than 0")
	}
	var err error
	for i := 0; i < amount; i++ {
		if err = x(); err != nil {
			time.Sleep(dur)
			continue
		}
		return nil
	}
	return err
}
