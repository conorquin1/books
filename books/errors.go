package books

import "errors"

var (
	// ErrBookNotFound is returned when a book is not found.
	ErrBookNotFound = errors.New("book not found")

	// ErrBookAlreadyExists is returned when trying to create a book that already exists.
	ErrBookAlreadyExists = errors.New("book already exists")

	// ErrInvalidBookData is returned when book data validation fails.
	ErrInvalidBookData = errors.New("invalid book data")
)

