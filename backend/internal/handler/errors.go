package handler

import "github.com/gigmatch/gigmatch/internal/constants"

func forbiddenErr() error {
	return constants.ErrForbidden
}
