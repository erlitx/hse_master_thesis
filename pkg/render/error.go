package render

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Err struct {
	Error string `json:"error"`
}

func Error(w http.ResponseWriter, err error, status int, message string) {
	log.Error().Err(err).Msg(message)

	err = unpack(err)
	err = fmt.Errorf("%s: %w", message, err)

	JSON(w, Err{Error: err.Error()}, status)
}

func unpack(err error) error {
	for {
		e := errors.Unwrap(err)
		if e == nil {
			break
		}

		err = e
	}

	return err
}

// ErrorSpecific maps domain errors to appropriate HTTP status codes and messages
func ErrorSpecific(w http.ResponseWriter, err error, domainErrors map[error]ErrorMapping) {
	status, message := mapError(err, domainErrors)

	log.Error().Err(err).Msg(message)

	err = unpack(err)
	err = fmt.Errorf("%s: %w", message, err)

	JSON(w, Err{Error: err.Error()}, status)
}

// ErrorMapping defines HTTP status and message for a domain error
type ErrorMapping struct {
	Status  int
	Message string
}

func mapError(err error, domainErrors map[error]ErrorMapping) (int, string) {
	for domainErr, mapping := range domainErrors {
		if errors.Is(err, domainErr) {
			return mapping.Status, mapping.Message
		}
	}

	// Default to internal server error
	return http.StatusInternalServerError, "internal server error"
}
