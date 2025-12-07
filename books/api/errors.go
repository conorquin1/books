package api

import (
	"log"
	"net/http"

	"github.com/books/books"
	"github.com/books/validate"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// getCodeByErr receives an error and returns its error code.
func getCodeByErr(err error) int {
	if _, ok := errors.Cause(err).(*validate.Validator); ok {
		return http.StatusBadRequest
	}

	switch newError := err; newError {
	case books.ErrInvalidBookData,
		books.ErrBookAlreadyExists:
		return http.StatusBadRequest
	case books.ErrBookNotFound:
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

// ErrorHandler is a middleware to handle errors in the api layer.
func ErrorHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err == nil {
			return nil
		}

		errCause := errors.Cause(err)
		code := getCodeByErr(errCause)
		if code == http.StatusInternalServerError {
			log.Printf("Error: %v", err)
		} else {
			log.Printf("Info: %v", err)
		}

		return echo.NewHTTPError(code, errCause.Error())
	}
}

