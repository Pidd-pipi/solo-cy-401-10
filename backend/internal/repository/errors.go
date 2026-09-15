// Package repository implements the data-access layer.
package repository

import "errors"

var (
	ErrNotFound = errors.New("resource not found")
	ErrConflict = errors.New("resource conflict")
)
