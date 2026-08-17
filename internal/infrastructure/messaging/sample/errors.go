package sample

import "errors"

func IsPoisonError(err error) bool {
	return errors.Is(err, ErrPoisonMessage)
}
